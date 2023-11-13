package cmd

import (
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
	"testing"
)

type validateArgTestCase struct {
	args     map[string]any
	expected string
}

var validateArgTestCases = map[string]validateArgTestCase{
	"host with no port": {
		args: map[string]any{
			constants.ArgHost: "http://localhost:7103",
		},
		// TODO kai update to use perr
		expected: "invalid 'host' argument: must be of form http://<host>:<port>",
	},
}

func TestValidateArgs(t *testing.T) {
	for name, testCase := range validateArgTestCases {
		for k, v := range testCase.args {
			viper.Set(k, v)
		}

		err := validateArgs()
		if err != nil {
			if err.Error() != testCase.expected {
				t.Errorf("test % failed: expected %s but got %s", name, testCase.expected, err.Error())
			}
		}
	}
}
