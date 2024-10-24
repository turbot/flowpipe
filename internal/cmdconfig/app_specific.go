package cmdconfig

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/app_specific_connection"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/modconfig/flowpipe"
)

// SetAppSpecificConstants sets app specific constants defined in pipe-fittings
func SetAppSpecificConstants() {
	app_specific.AppName = "flowpipe"

	versionString := viper.GetString("main.version")
	app_specific.AppVersion = semver.MustParse(versionString)

	// set all app specific env var keys
	app_specific.SetAppSpecificEnvVarKeys("FLOWPIPE_")

	app_specific.AutoVariablesExtensions = []string{".auto.fpvars"}
	app_specific.DefaultVarsFileName = "flowpipe.fpvars"
	app_specific.EnvInputVarPrefix = "FP_VAR_"

	app_specific.ConfigExtension = ".fpc"
	app_specific.ModDataExtensions = []string{".fp"}

	app_specific.VariablesExtensions = []string{".fpvars"}
	app_specific.WorkspaceIgnoreFile = ".flowpipeignore"
	app_specific.WorkspaceDataDir = ".flowpipe"

	// set the command pre and post hooks
	cmdconfig.CustomPreRunHook = preRunHook
	cmdconfig.CustomPostRunHook = postRunHook

	// Version check
	app_specific.VersionCheckHost = "hub.flowpipe.io"
	app_specific.VersionCheckPath = "api/cli/version/latest"

	// set the default install dir
	defaultInstallDir, err := files.Tildefy("~/.flowpipe")
	error_helpers.FailOnError(err)
	app_specific.DefaultInstallDir = defaultInstallDir

	// set the default config path
	globalConfigPath := filepath.Join(defaultInstallDir, "config")
	// check whether install-dir env has been set - if so, respect it
	if envInstallDir, ok := os.LookupEnv(app_specific.EnvInstallDir); ok {
		globalConfigPath = filepath.Join(envInstallDir, "config")
		app_specific.InstallDir = envInstallDir
	} else {
		/*
			NOTE:
			If InstallDir is settable outside of default & env var, need to add
			the following code to end of initGlobalConfig in init.go
			app_specific.InstallDir = viper.GetString(constants.ArgInstallDir) at end of
		*/
		app_specific.InstallDir = defaultInstallDir
	}
	app_specific.DefaultConfigPath = strings.Join([]string{".", globalConfigPath}, ":")

	// register supported connection types
	registerConnections()

	// set custom types
	var notifierImpl *flowpipe.NotifierImpl
	var notifierImplTypeName = reflect.TypeOf(notifierImpl).String()
	app_specific.CustomTypes = []string{notifierImplTypeName}

}

func registerConnections() {
	app_specific_connection.RegisterConnections(
		connection.NewAbuseIPDBConnection,
		connection.NewAlicloudConnection,
		connection.NewAwsConnection,
		connection.NewAzureConnection,
		connection.NewBitbucketConnection,
		connection.NewClickUpConnection,
		connection.NewDatadogConnection,
		connection.NewDiscordConnection,
		connection.NewFreshdeskConnection,
		connection.NewGcpConnection,
		connection.NewGithubConnection,
		connection.NewGitLabConnection,
		connection.NewIP2LocationIOConnection,
		connection.NewIPstackConnection,
		connection.NewJiraConnection,
		connection.NewJumpCloudConnection,
		connection.NewMastodonConnection,
		connection.NewMicrosoftTeamsConnection,
		connection.NewMysqlConnection,
		connection.NewDuckDbConnection,
		connection.NewOktaConnection,
		connection.NewOpenAIConnection,
		connection.NewOpsgenieConnection,
		connection.NewPagerDutyConnection,
		connection.NewPostgresConnection,
		connection.NewSendGridConnection,
		connection.NewServiceNowConnection,
		connection.NewSlackConnection,
		connection.NewSqliteConnection,
		connection.NewSteampipePgConnection,
		connection.NewTrelloConnection,
		connection.NewGuardrailsConnection,
		connection.NewPipesConnection,
		connection.NewUptimeRobotConnection,
		connection.NewUrlscanConnection,
		connection.NewVaultConnection,
		connection.NewVirusTotalConnection,
		connection.NewZendeskConnection)
}
