package flowpipe_invalid_tests

import (
	"context"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/parse"
	"github.com/turbot/flowpipe/internal/tests/test_init"
)

type testSetup struct {
	title         string
	file          string
	containsError string
	valid         bool
}

var tests = []testSetup{
	{
		title:         "invalid param definition",
		file:          "./pipelines/invalid_param_definition.fp",
		containsError: `A type specification is either a primitive type keyword (bool, number, string), complex type constructor call or Turbot custom type (connection, notifier)`,
	},
	{
		title:         "invalid param definition 2",
		file:          "./pipelines/invalid_param_definition_2.fp",
		containsError: `A type specification is either a primitive type keyword (bool, number, string), complex type constructor call or Turbot custom type (connection, notifier)`,
	},
	{
		title:         "bad output reference",
		file:          "./pipelines/bad_output_reference.fp",
		containsError: `invalid depends_on 'transform.does_not_exist' in output block, 'transform.does_not_exist' does not exist in pipeline local.pipeline.bad_output_reference`,
	},
	{
		title:         "duplicate pipeline",
		file:          "./pipelines/duplicate_pipelines.fp",
		containsError: "Mod defines more than one resource named 'local.pipeline.pipeline_007'",
	},
	{
		title:         "duplicate triggers - different pipeline",
		file:          "./pipelines/duplicate_triggers_diff_pipeline.fp",
		containsError: "Mod defines more than one resource named 'local.trigger.schedule.my_hourly_trigger'",
	},
	{
		title:         "duplicate triggers",
		file:          "./pipelines/duplicate_triggers.fp",
		containsError: "duplicate unresolved block name 'trigger.my_hourly_trigger'",
	},
	{
		title:         "invalid http trigger",
		file:          "./pipelines/invalid_http_trigger.fp",
		containsError: `Unsupported argument: An argument named "if" is not expected here.`,
	},
	{
		title:         "invalid schedule trigger - unsupported attribute 'execution_mode'",
		file:          "./pipelines/invalid_schedule_trigger.fp",
		containsError: `Unsupported argument: An argument named "execution_mode" is not expected here.`,
	},
	{
		title:         "invalid step attribute (transform)",
		file:          "./pipelines/invalid_step_attribute.fp",
		containsError: `Unsupported argument: An argument named "abc" is not expected here.`,
	},
	{
		title:         "invalid param",
		file:          "./pipelines/invalid_params.fp",
		containsError: `invalid property path: params.message_retention_duration`,
	},
	{
		title:         "invalid depends",
		file:          "./pipelines/invalid_depends.fp",
		containsError: "Failed to decode mod: invalid depends_on 'http.my_step_1', step 'http.my_step_1' does not exist in pipeline local.pipeline.invalid_depends",
	},
	{
		title:         "invalid email port",
		file:          "./pipelines/invalid_email_port.fp",
		containsError: "Failed to decode mod: Unable to convert port into integer\n",
	},
	{
		title:         "invalid email recipient",
		file:          "./pipelines/invalid_email_recipient.fp",
		containsError: "Unable to parse to attribute to string slice: Bad Request: expected string type, but got number\n",
	},
	{
		title:         "invalid container step attribute value - memory_swappiness",
		file:          "./pipelines/container_step_invalid_memory_swappiness.fp",
		containsError: "The value of 'memory_swappiness' attribute must be between 0 and 100",
	},
	{
		title:         "invalid trigger",
		file:          "./pipelines/invalid_trigger.fp",
		containsError: "Failed to decode mod: Missing required argument: The argument \"pipeline\" is required, but no definition was found.",
	},
	{
		title:         "invalid trigger - cron",
		file:          "./pipelines/invalid_trigger_bad_cron.fp",
		containsError: "bad cron format. Specify valid intervals hourly, daily, weekly, monthly or valid cron expression:",
	},
	{
		title:         "invalid loop - bad definition for transform step loop",
		file:          "./pipelines/loop_invalid_transform.fp",
		containsError: "Invalid attribute 'baz' in step loop block",
	},
	{
		title:         "invalid loop - bad definition for sleep step loop",
		file:          "./pipelines/loop_invalid_sleep.fp",
		containsError: "Invalid attribute 'baz' in the step loop block",
	},
	{
		title:         "invalid loop - no if",
		file:          "./pipelines/loop_no_if.fp",
		containsError: "The argument 'until' is required, but no definition was found",
	},
	{
		title:         "retry - multiple retry blocks",
		file:          "./pipelines/retry_multiple_retry_blocks.fp",
		containsError: "Only one retry block is allowed per step",
	},
	{
		title:         "retry - invalid attribute",
		file:          "./pipelines/retry_invalid_attribute.fp",
		containsError: "Unsupported attribute 'except' in retry block",
	},
	{
		title:         "retry - invalid attribute value",
		file:          "./pipelines/retry_invalid_attribute_value.fp",
		containsError: "Failed to decode mod: Unable to parse max_attempts attribute to integer\n(pipelines/retry_invalid_attribute_value.fp:7,13-33)",
	},
	{
		title:         "retry - invalid attribute value for strategy",
		file:          "./pipelines/retry_invalid_value_for_strategy.fp",
		containsError: "Invalid retry strategy: Valid values are constant, exponential or linear",
	},
	{
		title:         "throw - invalid attribute",
		file:          "./pipelines/throw_invalid_attribute.fp",
		containsError: "Unsupported argument 'foo' in throw block",
	},
	{
		title:         "throw - missing if",
		file:          "./pipelines/throw_missing_if.fp",
		containsError: "The argument 'if' is required",
	},
	{
		title:         "invalid pipeline output attribute - sensitive",
		file:          "./pipelines/output_invalid_attribute.fp",
		containsError: "Unsupported argument: An argument named \"sensitive\" is not expected here.",
	},
	{
		title:         "invalid error block attribute - ignored",
		file:          "./pipelines/invalid_error_attribute.fp",
		containsError: "Unsupported attribute 'ignored' in error block",
	},
	{
		title:         "invalid sleep step attribute - duration",
		file:          "./pipelines/invalid_sleep_attribute.fp",
		containsError: "Value of the attribute 'duration' must be a string or a whole number",
	},
	{
		title:         "invalid http step base attribute - timeout",
		file:          "./pipelines/invalid_http_timeout.fp",
		containsError: "Value of the attribute 'timeout' must be a string or a whole number",
	},
	{
		title:         "invalid schedule in query trigger",
		file:          "./pipelines/invalid_query_trigger.fp",
		containsError: "expected exactly 5 fields, found 1: [days]", // if not valid interval we assume it's a cron statement
	},
	{
		title:         "invalid execution mode in http trigger",
		file:          "./pipelines/invalid_http_trigger_execution_mode.fp",
		containsError: "The execution mode must be one of: synchronous,asynchronous",
	},
	// This test doesn't work because it needs FlowpipeConfig to load the notifier otherwise the notifier reference will break,
	// and notifier is a mandatory attribute so it will never test the option vs options
	// {
	// 	title:         "invalid input - option block(s) and options both set",
	// 	file:          "./pipelines/invalid_input_option_and_options.fp",
	// 	containsError: "Option blocks and options attribute are mutually exclusive",
	// },
	{
		title:         "invalid method types in http trigger",
		file:          "./pipelines/invalid_http_trigger_method.fp",
		containsError: "Method block type must be one of: post,get",
	},
	{
		title:         "duplicate method blocks in http trigger",
		file:          "./pipelines/invalid_http_trigger_duplicate_method.fp",
		containsError: "Duplicate method block for type: post",
	},
	{
		title:         "invalid query trigger - missing required field sql",
		file:          "./pipelines/query_trigger_missing_sql.fp",
		containsError: "The argument \"sql\" is required, but no definition was found.",
	},
	{
		title:         "invalid schedule trigger - missing required field schedule",
		file:          "./pipelines/schedule_trigger_missing_schedule.fp",
		containsError: "The argument \"schedule\" is required, but no definition was found.",
	},
	{
		title:         "duplicate output name",
		file:          "./pipelines/duplicate_output_name.fp",
		containsError: "duplicate output name 'output_test' - output names must be unique",
	},
	{
		title:         "invalid pipeline output attribute - depends_on",
		file:          "./pipelines/invalid_pipeline_output_attribute.fp",
		containsError: "Unsupported argument: An argument named \"depends_on\" is not expected here.",
	},
	{
		title:         "duplicate pipeline param name",
		file:          "./pipelines/duplicate_pipeline_param.fp",
		containsError: "duplicate pipeline parameter name 'my_param' - parameter names must be unique",
	},
	{
		title:         "default param value does not match type",
		file:          "./pipelines/param_default_mismatch.fp",
		containsError: "default value type mismatched - expected list of string, got string",
	},
	{
		title:         "default param value does not match type (2)",
		file:          "./pipelines/param_default_mismatch_2.fp",
		containsError: "default value type mismatched - expected set of bool, got number",
	},
	{
		title:         "invalid enum type, not a list",
		file:          "./pipelines/enum_param_invalid_enum_type.fp",
		containsError: "enum values must be a list",
	},
	{
		title:         "param enum must be scalar or list of scalar",
		file:          "./pipelines/enum_param_unsupported.fp",
		containsError: "enum is only supported for string, bool, number, list of string, list of bool, list of number types",
	},
	{
		title:         "enum does not match type - string vs list of number",
		file:          "./pipelines/enum_param_mismatched.fp",
		containsError: "enum values type mismatched",
	},
	{
		title:         "enum does not match type (2) - number vs list of string",
		file:          "./pipelines/enum_param_mismatched_2.fp",
		containsError: "enum values type mismatched",
	},
	{
		title:         "enum does not match type (3) - list of string vs list of number",
		file:          "./pipelines/enum_param_mismatched_3.fp",
		containsError: "enum values type mismatched",
	},
	{
		title:         "enum does not match type (4) - bool vs list of number",
		file:          "./pipelines/enum_param_mismatched_4.fp",
		containsError: "enum values type mismatched",
	},
	{
		title:         "enum param mismatched with default",
		file:          "./pipelines/enum_param_default_mismatched.fp",
		containsError: "enum is only supported for string, bool, number, list of string, list of bool",
	},
	{
		title: "enum param valid string",
		file:  "./pipelines/enum_param_valid_string.fp",
		valid: true,
	},
	{
		title: "enum param valid list of string",
		file:  "./pipelines/enum_param_valid_list_of_string.fp",
		valid: true,
	},
	{
		title: "enum param valid bool",
		file:  "./pipelines/enum_param_valid_bool.fp",
		valid: true,
	},
	{
		title: "enum param valid list of bool",
		file:  "./pipelines/enum_param_valid_list_of_bool.fp",
		valid: true,
	},
	{
		title: "enum param valid number",
		file:  "./pipelines/enum_param_valid_number.fp",
		valid: true,
	},
	{
		title: "enum param valid list of number",
		file:  "./pipelines/enum_param_valid_list_of_number.fp",
		valid: true,
	},
	{
		title:         "param default not in enum",
		file:          "./pipelines/enum_param_default_not_in_enum.fp",
		containsError: "default value not in enum",
	},
}

// Simple invalid test. Only single file resources can be evaluated here. This test is unable to test
// more complex error message expectations or complex structure such as mod & var
func TestSimpleInvalidResources(t *testing.T) {

	test_init.SetAppSpecificConstants()
	ctx := context.TODO()

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			assert := assert.New(t)

			if test.title == "" {
				assert.Fail("Test must contain a title")
				return
			}
			if !test.valid && test.containsError == "" {
				assert.Fail("Test: " + test.title + " does not have containsError test")
				return
			}

			_, _, err := parse.LoadPipelines(ctx, test.file)
			if test.valid && err == nil {
				return
			}

			if !test.valid && err == nil {
				assert.Fail("Test: " + test.title + " did not return an error")
				return
			}

			assert.Contains(err.Error(), test.containsError)

			// check that the error contains the filename
			assert.Contains(err.Error(), path.Base(test.file))
		})
	}

}
