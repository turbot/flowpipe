package gitleaks_fixture_ignore_tests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Places the secret scan reports that are fixtures or public samples, not credentials: the
// rule and the lines it reports in each file, as of the history the scan reads. The test checks
// the allowlist holds these places; it does not run gitleaks.
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

func TestGitleaksIgnoreCoversFixtureHits(t *testing.T) {
	assert := assert.New(t)

	body, err := os.ReadFile("../../../.gitleaksignore")
	if !assert.Nil(err, "could not read .gitleaksignore") {
		return
	}
	entries := map[string]bool{}
	for _, line := range strings.Split(string(body), "\n") {
		entries[strings.TrimSpace(line)] = true
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
