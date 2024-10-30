package invalid_mod_tests

import (
	"context"
	"errors"
	"github.com/turbot/flowpipe/internal/flowpipeconfig"
	fpparse "github.com/turbot/flowpipe/internal/parse"
	"github.com/turbot/flowpipe/internal/tests/test_init"
	"os"
	"path"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/workspace"
)

type FlowpipeSimpleInvalidConfigTestSuite struct {
	suite.Suite
	SetupSuiteRunCount    int
	TearDownSuiteRunCount int
	ctx                   context.Context
}

func (suite *FlowpipeSimpleInvalidConfigTestSuite) SetupSuite() {

	err := os.Setenv("RUN_MODE", "TEST_ES")
	if err != nil {
		panic(err)
	}

	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// clear the output dir before each test
	outputPath := path.Join(cwd, "output")

	// Check if the directory exists
	_, err = os.Stat(outputPath)
	if !os.IsNotExist(err) {
		// Remove the directory and its contents
		err = os.RemoveAll(outputPath)
		if err != nil {
			panic(err)
		}

	}

	pipelineDirPath := path.Join(cwd, "pipelines")

	viper.GetViper().Set("pipeline.dir", pipelineDirPath)
	viper.GetViper().Set("output.dir", outputPath)
	viper.GetViper().Set("log.dir", outputPath)

	// Create a single, global context for the application
	ctx := context.Background()

	suite.ctx = ctx

	// set app specific constants
	test_init.SetAppSpecificConstants()

	suite.SetupSuiteRunCount++
}

// The TearDownSuite method will be run by testify once, at the very
// end of the testing suite, after all tests have been run.
func (suite *FlowpipeSimpleInvalidConfigTestSuite) TearDownSuite() {
	suite.TearDownSuiteRunCount++
}

type invalidConfigTestSetup struct {
	title             string
	modDir            string
	configDirs        []string
	containsError     string
	errorType         string
	ignoreConfigParse bool
}

var invalidConfigTests = []invalidConfigTestSetup{
	{
		title:         "Invalid (unsupported) credential type",
		modDir:        "",
		configDirs:    []string{"./mods/invalid_cred"},
		containsError: "Invalid credential type slacks",
	},
	{
		title:         "Invalid slack integration",
		modDir:        "",
		configDirs:    []string{"./mods/bad_slack_integration_token_webhook_url"},
		containsError: "Attributes token and webhook_url are mutually exclusive",
	},
	{
		title:         "Invalid slack integration - missing both token and webhook_url",
		modDir:        "",
		configDirs:    []string{"./mods/bad_slack_integration_missing_token_webhook_url"},
		containsError: "slack.my_slack_app requires one of the following attributes set: token, webhook_url",
	},
	{
		title:         "Invalid slack integration - invalid signing_secret",
		modDir:        "",
		configDirs:    []string{"./mods/bad_slack_integration_signing_secret"},
		containsError: "Attribute signing_secret is only applies when attribute token is provided: slack.my_slack_app",
	},
	{
		title:         "Invalid email integration - missing required attributes: from, smtp_host",
		modDir:        "",
		configDirs:    []string{"./mods/bad_email_integration_missing_required_attribute"},
		containsError: "Missing required attributes from, smtp_host: email.my_email_app",
	},
	{
		title:         "Invalid email integration - missing required attribute: from",
		modDir:        "",
		configDirs:    []string{"./mods/bad_email_integration_missing_from"},
		containsError: "Attribute from must be defined: email.my_email_app",
	},
	{
		title:         "Invalid email integration - missing required attribute: smtp_host",
		modDir:        "",
		configDirs:    []string{"./mods/bad_email_integration_missing_smtp_host"},
		containsError: "Attribute smtp_host must be defined: email.my_email_app",
	},
	{
		title:         "Invalid email integration - invalid smtp_tls value",
		modDir:        "",
		configDirs:    []string{"./mods/bad_email_integration_smtp_tls"},
		containsError: "Attribute smtp_tls specified with invalid value dummy: email.my_email_app",
	},
	{
		title:         "Invalid notifier - no notify block provided",
		modDir:        "",
		configDirs:    []string{"./mods/bad_notifier_zero_notify"},
		containsError: "notifier must have at least one notify block to send the request to: admins",
	},
	{
		title:         "Invalid notify block - missing required attribute integration",
		modDir:        "",
		configDirs:    []string{"./mods/bad_notify_missing_integration"},
		containsError: "Missing required attribute: integration",
	},
	{
		title:         "Invalid notify block - invalid attribute 'cc' in slack integration",
		modDir:        "",
		configDirs:    []string{"./mods/bad_notify_unexpected_attribute_cc"},
		containsError: "Attribute 'cc' is not a valid attribute for slack type integration",
	},
	{
		title:         "Invalid notify block - invalid attribute 'bcc' in slack integration",
		modDir:        "",
		configDirs:    []string{"./mods/bad_notify_unexpected_attribute_bcc"},
		containsError: "Attribute 'bcc' is not a valid attribute for slack type integration",
	},
	{
		title:         "Invalid notify block - invalid attribute 'to' in slack integration",
		modDir:        "",
		configDirs:    []string{"./mods/bad_notify_unexpected_attribute_to"},
		containsError: "Attribute 'to' is not a valid attribute for slack type integration",
	},
	{
		title:         "Invalid notify block - invalid attribute 'channel' in slack integration",
		modDir:        "",
		configDirs:    []string{"./mods/bad_notify_unexpected_attribute_channel"},
		containsError: "Attribute 'channel' is not a valid attribute for email type integration",
	},
	{
		title:             "Duplicate message step",
		modDir:            "./mods/duplicate_message_step",
		configDirs:        []string{"./mods/duplicate_message_step"},
		ignoreConfigParse: true,
		containsError:     "duplicate step name 'message.test' - step names must be unique",
	},
	{
		title:             "Bad notifier reference to a string rather than an object",
		modDir:            "./mods/bad_notifier_reference",
		configDirs:        []string{"./mods/bad_notifier_reference"},
		ignoreConfigParse: true,
		containsError:     "Bad Request: notifier value must be a reference to a notifier resource",
	},
}

func (suite *FlowpipeSimpleInvalidConfigTestSuite) TestSimpleInvalidMods() {

	for _, test := range invalidConfigTests {

		suite.T().Run(test.title, func(t *testing.T) {
			assert := assert.New(t)
			if test.title == "" {
				assert.Fail("Test must have title")
				return
			}
			if test.containsError == "" {
				assert.Fail("Test " + test.title + " does not have containsError")
				return
			}

			fpConfig, errorAndWarning := flowpipeconfig.LoadFlowpipeConfig(test.configDirs)
			if errorAndWarning.Error == nil && !test.ignoreConfigParse {
				assert.FailNow("Expecting error but got nil")
				return
			}

			if !test.ignoreConfigParse {
				assert.Contains(errorAndWarning.Error.Error(), test.containsError)
			}
			notifierValueMap, err := fpConfig.NotifierValueMap()
			if err != nil {
				assert.Fail("Error getting notifier value map")
				return
			}
			if test.modDir != "" {
				_, errorAndWarning := workspace.Load(suite.ctx,
					test.modDir,
					workspace.WithConfigValueMap("notifier", notifierValueMap),
					workspace.WithDecoderOptions(fpparse.WithCredentials(fpConfig.Credentials)),
				)
				assert.NotNil(errorAndWarning.Error)
				if errorAndWarning.Error != nil {
					assert.Contains(errorAndWarning.Error.Error(), test.containsError)
				}

				if test.errorType != "" {
					var err perr.ErrorModel
					ok := errors.As(errorAndWarning.Error, &err)
					if !ok {
						assert.Fail("should be a pcerr.ErrorModel")
						return
					}

					assert.Equal(test.errorType, err.Type, "wrong error type")
				}
			}
		})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFlowpipeInvalidConfigTestSuite(t *testing.T) {
	suite.Run(t, new(FlowpipeSimpleInvalidConfigTestSuite))
}
