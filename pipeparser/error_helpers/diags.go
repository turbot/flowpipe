package error_helpers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/flowpipe/pipeparser/perr"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/terraform-components/tfdiags"
)

// DiagsToError converts tfdiags diags into an error
func DiagsToError(prefix string, diags tfdiags.Diagnostics) error {
	// convert the first diag into an error
	if !diags.HasErrors() {
		return nil
	}
	errorStrings := []string{prefix}
	// store list of messages (without the range) and use for deduping (we may get the same message for multiple ranges)
	errorMessages := []string{}
	for _, diag := range diags {
		if diag.Severity() == tfdiags.Error {
			errorString := diag.Description().Summary
			if diag.Description().Detail != "" {
				errorString += fmt.Sprintf(": %s", diag.Description().Detail)
			}

			if !helpers.StringSliceContains(errorMessages, errorString) {
				errorMessages = append(errorMessages, errorString)
				// now add in the subject and add to the output array
				if diag.Source().Subject != nil && len(diag.Source().Subject.Filename) > 0 {
					errorString += fmt.Sprintf("\n(%s)", diag.Source().Subject.StartString())
				}
				errorStrings = append(errorStrings, errorString)

			}
		}
	}
	if len(errorStrings) > 0 {
		errorString := strings.Join(errorStrings, "\n")
		if len(errorStrings) > 1 {
			errorString += "\n"
		}
		return errors.New(errorString)
	}
	return diags.Err()
}

func HclDiagsToError(prefix string, diags hcl.Diagnostics) error {
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
		return perr.InternalWithMessage(fmt.Sprintf("%s: %s", prefix, res))
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
