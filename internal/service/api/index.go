package api

import (
	"context"
	"fmt"
	"github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/types"
	"log/slog"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	ginlogger "github.com/FabienMht/ginslog/logger"
	ginrecovery "github.com/FabienMht/ginslog/recovery"
	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth/v7/limiter"
	"github.com/gin-contrib/gzip"
	size "github.com/gin-contrib/size"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	_ "github.com/swaggo/swag"
	"github.com/turbot/flowpipe/internal/log"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/service/api/middleware"
	"github.com/turbot/flowpipe/internal/service/api/service"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/utils"
)

// TODO - Test the changes

// @title Flowpipe
// @version 0.1.0
// @description Flowpipe is workflow and pipelines for DevSecOps.
// @contact.name Support
// @contact.email help@flowpipe.io

// @contact.name   Flowpipe
// @contact.url    http://www.flowpipe.io
// @contact.email  info@flowpipe.io

// @license.name  AGPLv3
// @license.url   https://www.gnu.org/licenses/agpl-3.0.en.html

// @host localhost
// @schemes https
// @BasePath /api/v0
// @query.collection.format multi

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// APIService represents the API service.
type APIService struct {
	// Ctx is the context used by the API service.
	ctx context.Context

	EsService *es.ESService

	httpServer *http.Server

	HTTPAddress string
	HTTPPort    int

	// Status tracking for the API service.
	Status    string
	StartedAt *time.Time

	StoppedAt      *time.Time `json:"stopped_at,omitempty"`
	apiPrefixGroup *gin.RouterGroup
	router         *gin.Engine
	ModMetadata    RootModMetadata
}

type RootModMetadata struct {
	IsStale    bool      `json:"is_stale"`
	LastLoaded time.Time `json:"last_loaded"`
}

// APIServiceOption defines a type of function to configures the APIService.
type APIServiceOption func(*APIService) error

// NewAPIService creates a new APIService.
func NewAPIService(ctx context.Context, es *es.ESService, opts ...APIServiceOption) (*APIService, error) {
	// Defaults
	api := &APIService{
		ctx:       ctx,
		EsService: es,
		Status:    "initialized",
		ModMetadata: RootModMetadata{
			IsStale:    false,
			LastLoaded: time.Now(),
		},
	}
	// Set options
	for _, opt := range opts {
		err := opt(api)
		if err != nil {
			return api, err
		}
	}
	return api, nil
}

// WithHTTPAddress sets the host and port of the API HTTPS service from the given
// address string in host:port format.
func WithHTTPAddress(addr string) APIServiceOption {
	return func(api *APIService) error {
		api.HTTPAddress = addr

		return nil
	}
}

// WithHTTPPort sets port of the API HTTP service
func WithHTTPPort(port int) APIServiceOption {
	return func(api *APIService) error {
		api.HTTPPort = port
		return nil
	}
}

// Start starts services managed by the Manager.
func (api *APIService) Start() error {

	slog.Debug("API starting")
	defer slog.Debug("API started")

	// Set the gin mode based on our environment, to configure logging etc as appropriate
	gin.SetMode(viper.GetString("environment"))
	binding.EnableDecoderDisallowUnknownFields = true

	// Initialize gin
	router := gin.New()

	// Add a ginslog middleware, which:
	//   - Logs all requests, like a combined access and error log.
	//   - Logs to stdout.
	//   - RFC3339 with UTC time format.
	router.Use(ginlogger.New(log.FlowpipeLogger()))

	// Logs all panic to error log
	//   - stack means whether output the stack info.
	router.Use(ginrecovery.New(log.FlowpipeLoggerWithLevelAndWriter(slog.LevelDebug, os.Stderr)))

	apiPrefixGroup := router.Group(common.APIPrefix())
	apiPrefixGroup.Use(common.ValidateAPIVersion)

	// Limit the size of POST requests
	// There doesn't seem a way to set the request size per path, but for now we have
	// no requirement for different limits on different paths. So just set one limit
	// for all request (for now)
	router.Use(size.RequestSizeLimiter(viper.GetInt64("web.request.size_limit")))

	// Create compression middleware - exclude process logs as we handle compression within the API itself
	compressionMiddleware := gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPathsRegexs([]string{"^/api/.+/.*[avatar|\\.jsonl]$"}))
	apiPrefixGroup.Use(compressionMiddleware)
	router.Use(compressionMiddleware)

	// Simple rate limiting:
	// * In memory only, so will not check across web servers
	// * Burst is the initial credits, with fill being added per second (to max of burst)
	//
	// Other option: ulele/limiter
	//
	// In the end decided to use tollbooth even though it doesn't have Redis support because that what was used in SPC
	// so I don't have to learn a new library.
	//
	// ulele/limiter support Redis AND in memory, so we may want to switch to that when we have more functionality in flowpipe
	//
	apiLimiter := tollbooth.NewLimiter(viper.GetFloat64("web.rate.fill"), &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour})
	apiLimiter.SetBurst(viper.GetInt("web.rate.burst"))
	router.Use(middleware.LimitHandler(apiLimiter))

	router.Use(middleware.SecurityMiddleware(api.ctx))

	service.RegisterPublicAPI(apiPrefixGroup)
	api.PipelineRegisterAPI(apiPrefixGroup)
	api.TriggerRegisterAPI(apiPrefixGroup)
	api.VariableRegisterAPI(apiPrefixGroup)
	api.ProcessRegisterAPI(apiPrefixGroup)
	api.DocsRegisterAPI(apiPrefixGroup)
	api.WebhookRegisterAPI(apiPrefixGroup)
	api.InputRegisterAPI(apiPrefixGroup)
	api.ModRegisterAPI(apiPrefixGroup)

	api.apiPrefixGroup = apiPrefixGroup
	api.router = router

	// Custom validators for our types
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// Return the JSON fieldname in the Tag() field for errors.
		// See https://github.com/go-playground/validator/issues/287
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
		// Custom validators using struct field tags
		_ = v.RegisterValidation("flowpipe_api_version", common.APIVersionValidator())
	}

	// Single Page App must catch all routes that are not found, it
	// handles them in a client side router.

	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method
		if strings.HasPrefix(path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": perr.NotFoundWithMessage(fmt.Sprintf("API Not Found: %s %s.", method, path))})
		} else {
			c.File("./static/index.html")
		}
	})

	// Server setup with graceful shutdown
	api.httpServer = &http.Server{
		Addr:              fmt.Sprintf("%s:%d", api.HTTPAddress, api.HTTPPort),
		Handler:           router,
		ReadHeaderTimeout: 60 * time.Second,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	// go func() {
	// 	if err := api.httpsServer.ListenAndServeTLS("./service/certificate/server.crt", "./service/certificate/server.key"); err != nil && err != http.ErrServerClosed {
	// slog.Error("API server failed to start", "error", err)
	// os.Exit(1)
	// 	}
	// }()

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := api.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("API server failed to start", "error", err)
			if output.IsServerMode {
				output.RenderServerOutput(api.ctx, types.NewServerOutputError(types.NewServerOutputPrefix(time.Now(), "flowpipe"), "API server failed to start", err))
			}
			os.Exit(1)
		}
	}()

	api.StartedAt = utils.TimeNow()
	api.Status = "running"

	return nil
}

// Stop stops services managed by the Manager.
func (api *APIService) Stop() error {
	slog.Debug("API stopping")
	defer slog.Debug("API stopped")

	// The context is used to inform the server it has time to finish the request
	// it is currently handling
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), time.Duration(viper.GetInt("web.server.cooldown_secs"))*time.Second)
	defer cancel()

	if api.httpServer != nil {
		if err := api.httpServer.Shutdown(ctxWithTimeout); err != nil {
			// TODO - wrap error
			return err
		}
		slog.Debug("API HTTP server stopped")
	}

	api.StoppedAt = utils.TimeNow()
	api.Status = "stopped"
	return nil
}
