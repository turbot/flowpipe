package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth/v7/limiter"
	"github.com/gin-contrib/gzip"
	size "github.com/gin-contrib/size"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/viper"
	_ "github.com/swaggo/swag"
	"github.com/zclconf/go-cty/cty"

	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/service/api/middleware"
	"github.com/turbot/flowpipe/internal/service/api/service"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/flowpipe/pipeparser/error_helpers"
	"github.com/turbot/flowpipe/pipeparser/funcs"
	"github.com/turbot/flowpipe/pipeparser/hclhelpers"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/perr"
	"github.com/turbot/flowpipe/pipeparser/utils"
)

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

	httpServer  *http.Server
	httpsServer *http.Server

	HTTPPort string `json:"http_port,omitempty"`

	HTTPSHost string `json:"https_host,omitempty"`
	HTTPSPort string `json:"https_port,omitempty"`

	// Status tracking for the API service.
	Status    string     `json:"status"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	StoppedAt *time.Time `json:"stopped_at,omitempty"`

	apiPrefixGroup *gin.RouterGroup
	router         *gin.Engine
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
		HTTPSHost: viper.GetString("web.https.host"),
		HTTPSPort: fmt.Sprintf("%d", viper.GetInt("web.https.port")),
		HTTPPort:  fmt.Sprintf("%d", viper.GetInt("web.http.port")),
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

// WithHTTPSAddress sets the host and port of the API HTTPS service from the given
// address string in host:port format.
func WithHTTPSAddress(addr string) APIServiceOption {
	return func(api *APIService) error {
		if addr == "" {
			return nil
		}
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return err
		}
		if host != "" {
			api.HTTPSHost = host
		}
		if port != "" {
			api.HTTPSPort = port
		}
		return nil
	}
}

func (api *APIService) RegisterHttpTriggers(triggers map[string]*modconfig.Trigger) error {

	salt, ok := cache.GetCache().Get("salt")
	if !ok {
		return perr.InternalWithMessage("salt not found")
	}

	routesInfo := api.router.Routes()

	currentTriggers := []string{}
	for _, routeInfo := range routesInfo {
		if routeInfo.Method == "POST" && strings.HasPrefix(routeInfo.Path, "/api/:api_version/hook/") {
			parts := strings.Split(routeInfo.Path, "/")
			triggerName := parts[len(parts)-2]
			hashString := parts[len(parts)-1]
			currentTriggers = append(currentTriggers, triggerName+"/"+hashString)
		}
	}

	// validTriggers := map[string]bool{}
	for _, t := range triggers {
		_, ok := t.Config.(*modconfig.TriggerHttp)
		if !ok {
			continue
		}

		inputString := t.FullName
		// Concatenate the input string and the salt
		concatenated := inputString + salt.(string)

		// Create a new SHA-256 hash
		hasher := sha256.New()

		// Write the concatenated string to the hasher
		hasher.Write([]byte(concatenated))

		// Get the final hash value
		hashBytes := hasher.Sum(nil)

		// Convert the hash to a hexadecimal string
		hashString := hex.EncodeToString(hashBytes)
		// validTriggers[t.FullName+"/"+hashString] = true

		// Check if we already have this route defined
		if slices.Contains[[]string, string](currentTriggers, t.FullName+"/"+hashString) {
			continue
		}

		api.apiPrefixGroup.POST("/hook/"+t.FullName+"/"+hashString, api.TriggerWebhook)
	}

	// There's no way to remove routes in Gin
	// now remove current triggers that are not in valid triggers
	// for _, t := range currentTriggers {
	// 	if !validTriggers[t] {
	// 		api.router.
	// 	}
	// }

	return nil
}

func parseURLPattern(urlPattern string) (string, string, error) {
	regexPattern := `^/api/v\d+/hook/([a-zA-Z0-9_\-\.]+)/([a-zA-Z0-9]+)`
	re := regexp.MustCompile(regexPattern)

	if !re.MatchString(urlPattern) {
		return "", "", perr.BadRequestWithMessage("Invalid URL pattern")
	}

	matches := re.FindStringSubmatch(urlPattern)

	if len(matches) != 3 {
		return "", "", perr.BadRequestWithMessage("Invalid URL pattern")
	}

	elements := strings.Split(matches[0], "/")

	if len(elements) != 6 {
		return "", "", perr.BadRequestWithMessage("Invalid URL pattern")
	}

	return matches[1], matches[2], nil
}

func (api *APIService) TriggerWebhook(c *gin.Context) {
	requestURL := c.Request.URL

	webhookTriggerName, webhookTriggerHash, err := parseURLPattern(requestURL.Path)

	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	// Get the trigger from the cache
	triggerCached, found := cache.GetCache().Get(webhookTriggerName)
	if !found {
		common.AbortWithError(c, perr.NotFoundWithMessage("trigger not found"))
		return
	}

	// check if the t is a webhook t
	t, ok := triggerCached.(*modconfig.Trigger)
	if !ok {
		common.AbortWithError(c, perr.NotFoundWithMessage("object is not a trigger"))
		return
	}

	_, ok = t.Config.(*modconfig.TriggerHttp)
	if !ok {
		common.AbortWithError(c, perr.NotFoundWithMessage("object is not a webhook trigger"))
		return
	}

	salt, ok := cache.GetCache().Get("salt")
	if !ok {
		common.AbortWithError(c, perr.InternalWithMessage("salt not found"))
		return
	}

	inputString := webhookTriggerName
	// Concatenate the input string and the salt
	concatenated := inputString + salt.(string)

	// Create a new SHA-256 hash
	hasher := sha256.New()

	// Write the concatenated string to the hasher
	hasher.Write([]byte(concatenated))

	// Get the final hash value
	hashBytes := hasher.Sum(nil)

	// Convert the hash to a hexadecimal string
	hashString := hex.EncodeToString(hashBytes)

	if hashString != webhookTriggerHash {
		common.AbortWithError(c, perr.UnauthorizedWithMessage("invalid hash"))
		return
	}

	// triggerRunnerHttp := trigger.TriggerRunnerHttp{
	// 	TriggerRunnerBase: trigger.TriggerRunnerBase{
	// 		Ctx:       api.ctx,
	// 		Trigger:   t,
	// 		EsService: api.EsService,
	// 	},
	// }

	body := ""
	if c.Request.Body != nil {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			common.AbortWithError(c, err)
			return
		}
		body = string(bodyBytes)
	}
	data := map[string]interface{}{}

	data["request_body"] = body
	// data["request_headers"] = c.Request.Header
	data["url"] = c.Request.RequestURI

	executionVariables := map[string]cty.Value{}

	selfObject := map[string]cty.Value{}
	for k, v := range data {
		ctyVal, err := hclhelpers.ConvertInterfaceToCtyValue(v)
		if err != nil {
			common.AbortWithError(c, err)
			return
		}
		selfObject[k] = ctyVal
	}

	executionVariables["self"] = cty.ObjectVal(selfObject)

	evalContext := &hcl.EvalContext{
		Variables: executionVariables,
		Functions: funcs.ContextFunctions(viper.GetString("work.dir")),
	}

	pipelineArgs, diags := t.GetArgs(evalContext)
	if diags.HasErrors() {
		common.AbortWithError(c, error_helpers.HclDiagsToError("trigger", diags))

	}

	pipeline := t.GetPipeline()
	pipelineName := pipeline.AsValueMap()["name"].AsString()

	pipelineCmd := &event.PipelineQueue{
		Event:               event.NewExecutionEvent(c),
		PipelineExecutionID: util.NewPipelineExecutionID(),
		Name:                pipelineName,
	}

	pipelineCmd.Args = pipelineArgs

	if err := api.EsService.Send(pipelineCmd); err != nil {
		common.AbortWithError(c, err)
		return
	}

	response := types.RunPipelineResponse{
		ExecutionID:           pipelineCmd.Event.ExecutionID,
		PipelineExecutionID:   pipelineCmd.PipelineExecutionID,
		ParentStepExecutionID: pipelineCmd.ParentStepExecutionID,
	}
	c.JSON(http.StatusOK, response)

}

// Start starts services managed by the Manager.
func (api *APIService) Start() error {

	// Convenience
	logger := fplog.Logger(api.ctx)

	logger.Debug("API starting")
	defer logger.Debug("API started")

	// Set the gin mode based on our environment, to configure logging etc as appropriate
	gin.SetMode(viper.GetString("environment"))

	// Initialize gin
	router := gin.New()

	// Add a ginzap middleware, which:
	//   - Logs all requests, like a combined access and error log.
	//   - Logs to stdout.
	//   - RFC3339 with UTC time format.
	router.Use(ginzap.Ginzap(logger.Zap, time.RFC3339, true))

	// Logs all panic to error log
	//   - stack means whether output the stack info.
	router.Use(ginzap.RecoveryWithZap(logger.Zap, true))

	// Set the same logger in all the Gin context
	router.Use(func(c *gin.Context) {
		c.Set("fplogger", logger)
	})

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
		Addr:              fmt.Sprintf("%s:%s", api.HTTPSHost, api.HTTPPort),
		Handler:           router,
		ReadHeaderTimeout: 60 * time.Second,
	}

	api.httpsServer = &http.Server{
		Addr:              fmt.Sprintf("%s:%s", api.HTTPSHost, api.HTTPSPort),
		Handler:           router,
		ReadHeaderTimeout: 60 * time.Second,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	// go func() {
	// 	if err := api.httpsServer.ListenAndServeTLS("./service/certificate/server.crt", "./service/certificate/server.key"); err != nil && err != http.ErrServerClosed {
	// 		log.Fatalf("listen: %s\n", err)
	// 	}
	// }()

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := api.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	api.StartedAt = utils.TimeNow()
	api.Status = "running"

	return nil
}

// Stop stops services managed by the Manager.
func (api *APIService) Stop() error {
	fplog.Logger(api.ctx).Debug("API stopping")
	defer fplog.Logger(api.ctx).Debug("API stopped")

	// The context is used to inform the server it has time to finish the request
	// it is currently handling
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), time.Duration(viper.GetInt("web.server.cooldown_secs"))*time.Second)
	defer cancel()

	if api.httpServer != nil {
		if err := api.httpServer.Shutdown(ctxWithTimeout); err != nil {
			// TODO - wrap error
			return err
		}
		fplog.Logger(api.ctx).Debug("API HTTP server stopped")
	}

	if api.httpsServer != nil {
		if err := api.httpsServer.Shutdown(ctxWithTimeout); err != nil {
			// TODO - wrap error
			return err
		}
		fplog.Logger(api.ctx).Debug("API HTTPS server stopped")
	}

	api.StoppedAt = utils.TimeNow()
	api.Status = "stopped"
	return nil
}
