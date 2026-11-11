package gitleaks_fixture_ignore_tests

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Places the secret scan reports that are fixtures or public samples, not credentials: the
// rule and the lines it reports in each file, as of the history the scan reads.
var fixtureHits = map[string]map[string][]int{
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
}

var entryPattern = regexp.MustCompile(`^[^\s:]+:[a-z0-9-]+:[1-9][0-9]*$`)

// parseIgnoreFile returns the entries of a .gitleaksignore body, skipping blank lines and
// comments. A line that is not file:rule:line is an error, since gitleaks would never match it.
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

func TestGitleaksIgnoreCoversFixtureHits(t *testing.T) {
	assert := assert.New(t)

	body, err := os.ReadFile("../../../.gitleaksignore")
	if !assert.Nil(err, "could not read .gitleaksignore") {
		return
	}
	entries, err := parseIgnoreFile(string(body))
	if !assert.Nil(err) {
		return
	}

	for path, rules := range fixtureHits {
		for rule, lines := range rules {
			for _, line := range lines {
				entry := fmt.Sprintf("%s:%s:%d", path, rule, line)
				assert.True(entries[entry], "missing .gitleaksignore entry %s", entry)
			}
		}
	}
}

func TestParseIgnoreFileRejectsMalformedLines(t *testing.T) {
	assert := assert.New(t)

	for _, bad := range []string{
		"service/certificate/server.key",                   // no rule or line
		"service/certificate/server.key:private-key",       // no line
		"service/certificate/server.key:private-key:0",     // line numbers start at 1
		"service/certificate/server.key:private-key:abc",   // not a number
		"service/certificate/*.key:private-key:1 # a note", // trailing text
	} {
		_, err := parseIgnoreFile(bad + "\n")
		assert.NotNil(err, "expected %q to be rejected", bad)
	}

	entries, err := parseIgnoreFile("# a comment\n\na/b.fpc:private-key:2\n")
	assert.Nil(err)
	assert.True(entries["a/b.fpc:private-key:2"])
	assert.Len(entries, 1)
}
