package config

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/spf13/viper"

	"github.com/turbot/flowpipe/internal/service/api/common"
)

var ConfigDefaults = map[string]any{
	// Set to debug or release - default to debug and envs override this to release as required
	"environment": "debug",

	// Set to single or cluster - default to cluster and envs override this to single as required
	"redis.mode": "redis",

	// Rate limiting defaults to an initial/max balance of 100, refilling
	// at 10 per second.
	"web.rate.fill":     10,
	"web.rate.burst":    100,
	"api.rate.interval": 10,
	"api.rate.limit":    100,

	// POST requests
	"web.request.size_limit": 1 * 1024 * 1024,

	// Cooldown time after SIGINT etc to allow existing requests to finish
	"web.server.cooldown_secs": 5,

	// Assume development mode by default, using localhost
	"web.secure.allowed_hosts": []string{"localhost"},
	"web.secure.ssl_host":      "localhost:7103",

	// Paging limit default and boundaries
	"api.list.limit.default": 25,
	"api.list.limit.min":     1,
	"api.list.limit.max":     100,

	// API user permissions
	"api.user.permissions.ttl": 600,

	// The bcrypt cost to use when hashing the token. defaults to 10, which is
	// the current bcrypt.DefaultCost in https://pkg.go.dev/golang.org/x/crypto/bcrypt#pkg-constants
	// Use at least 12 per https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#bcrypt
	"secrets.token.cost": 12,

	// AWS IAM roles can configure their max session between 1hr and 12 hours. The default max is 1 hr.
	// Then, when assuming a role, the minimum you can request is 15 mins. The default is 1hr for assuming rides, and 12hr in identity center stuff.
	// So as a starting point - min 15 mins, max 12 hrs and default 1hr.
	// System User session parameters
	"api.auth.jwt.expiration_seconds.system":     60 * 60,      // default - 1 hour
	"api.auth.jwt.expiration_seconds.min.system": 15 * 60,      // min - 15 minutes
	"api.auth.jwt.expiration_seconds.max.system": 60 * 60 * 12, // max - 12 hours

	// Other User session parameters
	"api.auth.jwt.expiration_seconds.user":            60 * 60 * 24 * 30, // 30 days
	"api.auth.jwt.expiration_seconds.min.user":        60 * 60,           // 1 hour
	"api.auth.jwt.expiration_seconds.max.user":        60 * 60 * 24 * 30, // 30 days
	"api.auth.login.token.request.expiration_seconds": 60 * 15,           // 15 minutes

	// Workspace auth
	"api.workspace.jwt.signing.method":          "RS256",
	"api.workspace.snapshot.jwt.signing.method": "HS256",
}

func Initialize() {

	//
	// CONFIG SETUP
	//

	baseUrl := viper.GetString("url.base")
	u, err := url.Parse(baseUrl)
	if err != nil {
		log.Fatal("Unable to parse url.base", err)
	}

	viper.Set("domain.base", u.Hostname())
	viper.Set("api.url.base", fmt.Sprintf("%s%s", baseUrl, common.PathPrefixWithVersion(common.APIVersionLatest)))

	// TODO KAI what is this used for and does it need to fit in with existing config
	// Allow environment variables for all config options.
	// e.g. session.secret becomes STEAMPIPE_CLOUD_SESSION_SECRET
	viper.AutomaticEnv()
	viper.SetEnvPrefix(viper.GetString("advanced.env_var_prefix"))

	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
}
