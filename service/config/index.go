package config

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/spf13/viper"

	"github.com/turbot/flowpipe/service/api/common"
)

// Initialize the configuration from steampipe.yml
func Initialize(_ context.Context) {

	//
	// DEFAULTS
	//

	// Set to debug or release - default to debug and envs override this to release as required
	viper.SetDefault("environment", "debug")
	viper.SetDefault("gcp.project", "localhost")

	viper.SetDefault("shard.default.region", "apse1")
	viper.SetDefault("shard.default.shard", "0001")

	viper.SetDefault("url.base", "https://localhost:3000")
	viper.SetDefault("db.url.base", "localhost")
	viper.SetDefault("dashboard.url.base", "localhost")

	viper.SetDefault("web.http.port", 7102)
	viper.SetDefault("web.https.port", 7103)

	viper.SetDefault("snapshot.render.http.port", 9196)
	viper.SetDefault("snapshot.render.https.port", 9197)

	// Set to single or cluster - default to cluster and envs override this to single as required
	viper.SetDefault("redis.mode", "redis")

	// Analytics
	viper.SetDefault("api.analytics.heap.app_id", "2186332868")

	viper.SetDefault("api.admin.user_limit", 1000)

	// Rate limiting defaults to an initial/max balance of 100, refilling
	// at 10 per second.
	viper.SetDefault("web.rate.fill", 10)
	viper.SetDefault("web.rate.burst", 100)
	viper.SetDefault("api.rate.interval", 10)
	viper.SetDefault("api.rate.limit", 100)

	// POST requests
	viper.SetDefault("web.request.size_limit", 1*1024*1024)

	// Cooldown time after SIGINT etc to allow existing requests to finish
	viper.SetDefault("web.server.cooldown_secs", 5)

	// Assume development mode by default, using localhost
	viper.SetDefault("web.secure.allowed_hosts", []string{"localhost"})
	viper.SetDefault("web.secure.ssl_host", "localhost:8443")

	// Paging limit default and boundaries
	viper.SetDefault("api.list.limit.default", 25)
	viper.SetDefault("api.list.limit.min", 1)
	viper.SetDefault("api.list.limit.max", 100)

	// API user permissions
	viper.SetDefault("api.user.permissions.ttl", 600)
	viper.SetDefault("api.experimental.aggregator", false)
	viper.SetDefault("api.experimental.billing", false)
	viper.SetDefault("api.experimental.connection", false)
	viper.SetDefault("api.experimental.dashboard", false)
	viper.SetDefault("api.experimental.datatank", false)
	viper.SetDefault("api.experimental.export", false)
	viper.SetDefault("api.experimental.notification", false)
	viper.SetDefault("api.experimental.pipeline", false)
	viper.SetDefault("api.experimental.process", false)
	viper.SetDefault("api.workspace.experimental.prometheus", false)

	// API Files
	viper.SetDefault("api.file.bucket", "steampipe-cloud-latest-snapshots-apse1")
	viper.SetDefault("api.file.provider", "local")

	// API Snapshots
	viper.SetDefault("api.workspace.snapshot.bucket", "steampipe-cloud-latest-snapshots-apse1")
	viper.SetDefault("api.workspace.snapshot.provider", "local")

	viper.SetDefault("api.workspace.processLogs.provider", "gcs")

	// The bcrypt cost to use when hashing the token. defaults to 10, which is
	// the current bcrypt.DefaultCost in https://pkg.go.dev/golang.org/x/crypto/bcrypt#pkg-constants
	// Use at least 12 per https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#bcrypt
	viper.SetDefault("secrets.token.cost", 12)

	// Authentication session parameters
	viper.SetDefault("session.secret", "upbeat_tesla")
	viper.SetDefault("api.auth.cookie.name", "cloud_token")
	viper.SetDefault("api.auth.cookie.path", "/")
	viper.SetDefault("api.auth.cookie.http_only", true)
	viper.SetDefault("api.auth.cookie.secure", true)
	viper.SetDefault("api.auth.jwt.signing.method", "HS256")

	// AWS IAM roles can configure their max session between 1hr and 12 hours. The default max is 1 hr.
	// Then, when assuming a role, the minimum you can request is 15 mins. The default is 1hr for assuming rides, and 12hr in identity center stuff.
	// So as a starting point - min 15 mins, max 12 hrs and default 1hr.
	// System User session parameters
	viper.SetDefault("api.auth.jwt.expiration_seconds.system", 60*60)        // default - 1 hour
	viper.SetDefault("api.auth.jwt.expiration_seconds.min.system", 15*60)    // min - 15 minutes
	viper.SetDefault("api.auth.jwt.expiration_seconds.max.system", 60*60*12) // max - 12 hours

	// Other User session parameters
	viper.SetDefault("api.auth.jwt.expiration_seconds.user", 60*60*24*30)      // 30 days
	viper.SetDefault("api.auth.jwt.expiration_seconds.min.user", 60*60)        // 1 hour
	viper.SetDefault("api.auth.jwt.expiration_seconds.max.user", 60*60*24*30)  // 30 days
	viper.SetDefault("api.auth.login.token.request.expiration_seconds", 60*15) // 15 minutes

	// Workspace auth
	viper.SetDefault("api.workspace.jwt.signing.method", "RS256")
	viper.SetDefault("api.workspace.snapshot.jwt.signing.method", "HS256")

	// Worker tasks
	viper.SetDefault("task.queue.max_size", 100) // Max number of items in a worker queue
	viper.SetDefault("task.workers.max", 1)      // Max number of workers

	viper.SetDefault("worker.k8s.log.level", "default")
	viper.SetDefault("worker.cleanup.connection.limit", 100)
	viper.SetDefault("worker.cleanup.workspace.limit", 10)
	viper.SetDefault("worker.cleanup.org.limit", 10)
	viper.SetDefault("worker.cleanup.user.limit", 10)
	viper.SetDefault("worker.cleanup.pipeline.limit", 10)
	viper.SetDefault("worker.cleanup.db.log.delete.rate", 1000)
	viper.SetDefault("worker.cleanup.process.archive.rate", 100)
	viper.SetDefault("worker.cleanup.process.log.limit", 200)
	viper.SetDefault("worker.cleanup.process.log.archive.hour", 1)
	viper.SetDefault("worker.cleanup.process.log.archive.rate", 100)

	viper.SetDefault("worker.db.log.heap", false)

	// Integrations
	viper.SetDefault("api.integration.github.app_id", "104915")
	viper.SetDefault("api.integration.github.client_id", "7519fa097ef9950eaf05")
	viper.SetDefault("api.integration.gitlab.app_id", "609c2b16a10aebb283b42b65bfdffb931873a60fd4ea4e1e0d715d9e1d7ed4c5")
	viper.SetDefault("api.integration.google.client_id", "1089654030488-bu2i6rekd1mp9vpfuc46jm9nu08ahuar.apps.googleusercontent.com")

	// Advanced options
	// Prefix to use for env var configuration, e.g. session.secret becomes STEAMPIPE_CLOUD_SESSION_SECRET
	viper.SetDefault("advanced.env_var_prefix", "STEAMPIPE_CLOUD")

	// Steampipe Container setup
	viper.SetDefault("steampipe.safeToEvict", "false")
	viper.SetDefault("steampipe.log.level", "warn")

	viper.SetDefault("envoy.version", "v1.15-latest")

	// The last 4 digit of a cron schedule. The minute schedule will be automatically
	// generated so 1 here is the hour schedule.
	viper.SetDefault("steampipe.update.schedule", "2 * * *")
	viper.SetDefault("steampipe.update.max_hour", 12)

	// Vault related config
	viper.SetDefault("vault.token.path", "/tmp/vault-token")
	viper.SetDefault("vault.address", "http://127.0.0.1:8200")

	viper.SetDefault("temporal.enable", true)
	viper.SetDefault("temporal.host", "localhost")
	viper.SetDefault("temporal.port", "7233")

	viper.SetDefault("awsTransit.oidc.path", "/opt/steampipe/oidc/token")

	viper.SetDefault("temporal.cacrt", "/opt/steampipe/temporal/certs/ca.crt")
	viper.SetDefault("temporal.tlscrt", "/opt/steampipe/temporal/certs/tls.crt")
	viper.SetDefault("temporal.tlskey", "/opt/steampipe/temporal/certs/tls.key")

	viper.SetDefault("k8s.auth-type", "internal")

	viper.SetDefault("monitor.workspace.email", "admin@turbot.com")
	viper.SetDefault("monitor.api.email", "admin@turbot.com")

	// Billing
	viper.SetDefault("billing.trial_period.days", 14)
	viper.SetDefault("billing.suspend.grace_period.days", 14)
	viper.SetDefault("billing.delete.grace_period.days", 30)

	//
	// CONFIG SETUP
	//

	// Gather configuration from the steampipe.yaml file in the current
	// directory
	viper.SetConfigName("steampipe.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/opt/steampipe/conf")

	/*
		// Read in the steampipe.yml file
		if err := viper.ReadInConfig(); err != nil {
			log.Fatal("steampipe.yml config file not found", err)
		}
	*/

	baseUrl := viper.GetString("url.base")
	u, err := url.Parse(baseUrl)
	if err != nil {
		log.Fatal("Unable to parse url.base", err)
	}

	viper.Set("domain.base", u.Hostname())
	viper.Set("api.url.base", fmt.Sprintf("%s%s", baseUrl, common.PathPrefixWithVersion(common.APIVersionLatest)))

	// Allow environment variables for all config options.
	// e.g. session.secret becomes STEAMPIPE_CLOUD_SESSION_SECRET
	viper.AutomaticEnv()
	viper.SetEnvPrefix(viper.GetString("advanced.env_var_prefix"))
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

}
