package test_init

import (
	fparse "github.com/turbot/flowpipe/internal/parse"
	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/app_specific_connection"
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/parse"
	"github.com/zclconf/go-cty/cty"
	"reflect"
)

func SetAppSpecificConstants() {
	installDir, err := files.Tildefy("~/.flowpipe")
	if err != nil {
		panic(err)
	}

	app_specific.AppName = "flowpipe"
	// TODO unify version logic with steampipe
	//app_specific.AppVersion
	app_specific.AutoVariablesExtensions = []string{".auto.fpvars"}
	//app_specific.ClientConnectionAppNamePrefix
	//app_specific.ClientSystemConnectionAppNamePrefix
	app_specific.DefaultInstallDir = installDir
	app_specific.DefaultVarsFileName = "flowpipe.fpvars"
	//app_specific.DefaultDatabase
	//app_specific.EnvAppPrefix
	app_specific.EnvInputVarPrefix = "FP_VAR_"
	//app_specific.InstallDir
	app_specific.ConfigExtension = ".fpc"
	app_specific.ModDataExtensions = []string{".fp"}
	app_specific.VariablesExtensions = []string{".fpvars"}
	//app_specific.ServiceConnectionAppNamePrefix
	app_specific.WorkspaceIgnoreFile = ".flowpipeignore"
	app_specific.WorkspaceDataDir = ".flowpipe"

	// set custom types
	app_specific.CustomTypes = map[string]cty.Type{"notifier": cty.Capsule("BaseNotifierCtyType", reflect.TypeOf(&resources.NotifierImpl{}))}

	// set parse related constants
	parse.ModDecoderFunc = fparse.NewFlowpipeModDecoder
	modconfig.AppSpecificNewModResourcesFunc = resources.NewModResources

	app_specific_connection.RegisterConnections(
		connection.NewAbuseIPDBConnection,
		connection.NewAlicloudConnection,
		connection.NewAwsConnection,
		connection.NewAzureConnection,
		connection.NewBitbucketConnection,
		connection.NewClickUpConnection,
		connection.NewDatadogConnection,
		connection.NewDiscordConnection,
		connection.NewDuckDbConnection,
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
