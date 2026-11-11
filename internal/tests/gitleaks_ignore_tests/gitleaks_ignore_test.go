package gitleaks_ignore_tests

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The aws-access-token matches the secret scan reports in the test fixtures (issue #1135).
// Each path maps to the lines the scan reports in it, as of the history it reads.
var awsFixtureLines = map[string][]int{
	"internal/tests/flowpipe_mod_tests/mod_with_creds_output/creds.fpc": {2},
	"internal/tests/flowpipe_mod_tests/mod_with_conn_output/creds.fpc":  {2},
	"internal/tests/flowpipe_mod_tests/custom_type_three/conn.fpc":      {2, 7, 12},
	"internal/tests/flowpipe_mod_tests/custom_type/conn.fpc":            {2, 7, 12},
	"internal/tests/flowpipe_mod_tests/custom_type_two/conn.fpc":        {2, 7, 12},
	"internal/tests/flowpipe_mod_tests/custom_type_notifier/conn.fpc":   {2, 7, 12},
	"internal/tests/flowpipe_mod_tests/custom_type_four/conn.fpc":       {2, 7, 12},
	"internal/tests/flowpipe_parsing_tests/pipelines/credentials.fp":    {3},
	"internal/es/estest/mod_two_test.go":                                {828, 853, 880, 907, 934, 960},
	"internal/tests/flowpipe_mod_tests/flowpipe_mod_suite_test.go":      {605, 640},
	"internal/es/estest/test_suite_mod_2/conn.fpc":                      {2, 7, 12},
	"internal/es/estest/default_mod_test.go":                            {154, 200, 516, 562, 632},
	"internal/es/estest/default_mod/connection.fpc":                     {2},
	"internal/es/estest/default_mod/credential.fpc":                     {2},
	"scripts/sanity_check_mod/credential.fpc":                           {2},
	"internal/es/estest/test_suite_mod/.flowpipe/config/workspaces.fpc": {2},
	"internal/es/estest/test_suite_mod/pipelines/credentials.fp":        {3},
	"internal/es/estest/test_suite_mod/creds.fpc":                       {2},
	"internal/es/estest/mod_test.go":                                    {2008, 2013},
}

// Other places the scan reports that are fixtures or public samples, not credentials: the rule
// and the lines it reports in each file, as of the history the scan reads.
var otherFixtureHits = map[string]map[string][]int{
	"internal/tests/flowpipe_mod_tests/steampipe_config/gitlab.spc":     {"generic-api-key": {4, 10}},
	"internal/tests/flowpipe_mod_tests/steampipe_config/ipstack.spc":    {"generic-api-key": {3, 8}},
	"internal/tests/flowpipe_mod_tests/steampipe_config/guardrails.spc": {"generic-api-key": {4, 5, 12, 13}},
	"internal/tests/flowpipe_mod_tests/steampipe_config/zendesk.spc":    {"generic-api-key": {6, 14}},
	"internal/tests/flowpipe_mod_tests/steampipe_config/datadog.spc":    {"generic-api-key": {4, 6, 12, 14}},
	"internal/es/estest/mod_test.go":                                    {"generic-api-key": {2117, 2134}, "gitlab-pat": {2193}},
	"internal/es/estest/test_suite_mod/.flowpipe/config/workspaces.fpc": {"gitlab-pat": {24}, "generic-api-key": {24}},
	"pipeparser/lang/funcs/crypto_test.go":                              {"private-key": {687, 716, 745, 774}},
	"service/certificate/server.key":                                    {"private-key": {1}},
	"internal/sanitize/sanitizer_test.go":                               {"github-pat": {35}},
	"internal/es/estest/test_suite_mod/pipelines/input_slack.fp":        {"generic-api-key": {2}},
	"internal/es/test/test_suite_mod/pipelines/input_slack.hcl":         {"generic-api-key": {2}},
}

var entryPattern = regexp.MustCompile(`^[^\s:]+:[a-z0-9-]+:[1-9][0-9]*$`)

// parseIgnoreFile returns the entries of a .gitleaksignore body, skipping blank lines and
// comments. A line that is not file:rule:line is an error: this file lists entries in that form
// only, although gitleaks also accepts commit:file:rule:line.
func parseIgnoreFile(body string) (map[string]bool, error) {
	entries := map[string]bool{}
	scanner := bufio.NewScanner(strings.NewReader(body))
	for n := 1; scanner.Scan(); n++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !entryPattern.MatchString(line) {
			return nil, fmt.Errorf("line %d is not file:rule:line: %q", n, line)
		}
		entries[line] = true
	}
	return entries, scanner.Err()
}

func TestGitleaksIgnoreCoversAwsFixtures(t *testing.T) {
	assert := assert.New(t)

	body, err := os.ReadFile("../../../.gitleaksignore")
	if !assert.Nil(err, "could not read .gitleaksignore") {
		return
	}

	entries, err := parseIgnoreFile(string(body))
	if !assert.Nil(err) {
		return
	}

	for path, lines := range awsFixtureLines {
		for _, line := range lines {
			entry := fmt.Sprintf("%s:aws-access-token:%d", path, line)
			assert.True(entries[entry], "missing .gitleaksignore entry %s", entry)
		}
	}
}

func TestGitleaksIgnoreCoversOtherFixtureHits(t *testing.T) {
	assert := assert.New(t)

	body, err := os.ReadFile("../../../.gitleaksignore")
	if !assert.Nil(err, "could not read .gitleaksignore") {
		return
	}

	entries, err := parseIgnoreFile(string(body))
	if !assert.Nil(err) {
		return
	}

	for path, rules := range otherFixtureHits {
		for rule, lines := range rules {
			for _, line := range lines {
				entry := fmt.Sprintf("%s:%s:%d", path, rule, line)
				assert.True(entries[entry], "missing .gitleaksignore entry %s", entry)
			}
		}
	}
}

func TestGitleaksIgnoreOnlyAllowsListedFixtures(t *testing.T) {
	assert := assert.New(t)

	body, err := os.ReadFile("../../../.gitleaksignore")
	if !assert.Nil(err, "could not read .gitleaksignore") {
		return
	}

	entries, err := parseIgnoreFile(string(body))
	if !assert.Nil(err) {
		return
	}

	// Every entry must be one of the listed fixture places: no wildcard, no whole-file or
	// whole-rule allowance that would hide a real leak.
	allowed := map[string]bool{}
	for path, lines := range awsFixtureLines {
		for _, line := range lines {
			allowed[fmt.Sprintf("%s:aws-access-token:%d", path, line)] = true
		}
	}
	for path, rules := range otherFixtureHits {
		for rule, lines := range rules {
			for _, line := range lines {
				allowed[fmt.Sprintf("%s:%s:%d", path, rule, line)] = true
			}
		}
	}
	for entry := range entries {
		assert.True(allowed[entry], "unexpected .gitleaksignore entry %s", entry)
	}
}

func TestParseIgnoreFileRejectsMalformedLines(t *testing.T) {
	assert := assert.New(t)

	for _, bad := range []string{
		"internal/es/estest/mod_test.go",                      // no rule or line
		"internal/es/estest/mod_test.go:aws-access-token",     // no line
		"internal/es/estest/mod_test.go:aws-access-token:0",   // line numbers start at 1
		"internal/es/estest/mod_test.go:aws-access-token:abc", // not a number
		"internal/es/estest/*.go:aws-access-token:1 # note",   // trailing text
	} {
		_, err := parseIgnoreFile(bad + "\n")
		assert.NotNil(err, "expected %q to be rejected", bad)
	}

	entries, err := parseIgnoreFile("# a comment\n\na/b.fpc:aws-access-token:2\n")
	assert.Nil(err)
	assert.True(entries["a/b.fpc:aws-access-token:2"])
	assert.Len(entries, 1)
}
