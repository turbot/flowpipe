package pipeparser

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/flowpipe/fperr"
)

// DiagsToError converts tfdiags diags into an error
func DiagsToError(prefix string, diags hcl.Diagnostics) error {
	if !diags.HasErrors() {
		return nil
	}
	errStrings := diagsToString(diags, hcl.DiagError)

	var res string
	if len(errStrings) > 0 {
		res = strings.Join(errStrings, "\n")
		if len(errStrings) > 1 {
			res += "\n"
		}
		return fperr.InternalWithMessage(fmt.Sprintf("%s: %s", prefix, res))
	}

	return diags.Errs()[0]
}

func diagsToString(diags hcl.Diagnostics, severity hcl.DiagnosticSeverity) []string { // convert the first diag into an error
	// store list of messages (without the range) and use for de-duping (we may get the same message for multiple ranges)
	var msgMap = make(map[string]struct{})
	var strs []string
	for _, diag := range diags {
		if diag.Severity == severity {
			str := diag.Summary
			if diag.Detail != "" {
				str += fmt.Sprintf(": %s", diag.Detail)
			}

			if _, ok := msgMap[str]; !ok {
				msgMap[str] = struct{}{}
				// now add in the subject and add to the output array
				if diag.Subject != nil && len(diag.Subject.Filename) > 0 {
					str += fmt.Sprintf("\n(%s)", diag.Subject.String())
				}

				strs = append(strs, str)
			}
		}
	}

	return strs
}

// LoadFileData builds a map of filepath to file data
func LoadFileData(paths ...string) (map[string][]byte, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var fileData = map[string][]byte{}

	for _, configPath := range paths {
		data, err := os.ReadFile(configPath)

		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("failed to read config file %s", configPath),
				Detail:   err.Error()})
			continue
		}
		fileData[configPath] = data
	}
	return fileData, diags
}
