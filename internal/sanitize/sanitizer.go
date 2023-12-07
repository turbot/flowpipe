package sanitize

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"slices"
	"sort"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-code/secrets"
)

const RedactedStr = "REDACTED"

var NullSanitizer = NewSanitizer(SanitizerOptions{})

var Instance = NewSanitizer(SanitizerOptions{
	ExcludeFields: []string{
		"password",
		"secretaccesskey",
		"sessiontoken",
		"smtp_password",
		"api_key",
		"api_token",
		// "key", // we can't sanitize key because of each.key
		"token",
		"cloud_token",
		"aws_access_key_id",
		"aws_secret_access_key",
		"aws_session_token",
		"clientsecret",
		"access_token",
		"access_key",
		"sourcerecord",
		"cert",
		"privatekey",
		"secretvalue",
		"slack_token",
		"abuseipdb_api_key",
		"sendgrid_api_key",
		"vtcli_apikey",
		"zendesk_token",
		"trello_api_key",
		"trello_token",
		"okta_token",
		"uptimerobot_api_key",
		"urlscan_api_key",
		"clickup_token",
		"pagerduty_token",
		"discord_token",
		"ip2location_api_key",
		"ipstack_access_key",
		"teams_access_token",
		"pipes_token",
		"github_token",
		"gitlab_token",
		"vault_token",
		"jira_api_token",
	},
	ExcludePatterns: []string{
		`SG\.[a-zA-Z0-9_-]{22}\.[a-zA-Z0-9_-]{43}`, // sendgrid
		`AccountKey=[a-zA-Z0-9+/=]{88}`,            // azure storage account key
	},
	ImportCodeMatchers: true,
})

type SanitizerOptions struct {
	// ExcludeFields is a list of fields to exclude from sanitization
	ExcludeFields []string
	// ExcludePatterns is a list of regexes - any capture groups are redacted
	ExcludePatterns []string

	ImportCodeMatchers bool
}

type Sanitizer struct {
	regexes             []*regexp.Regexp
	fieldPatternRegexes []*regexp.Regexp
	excludeFields       map[string]struct{}
}

var codePluginExcludedRegex = []string{
	"heroku_api_key",
	"azure_storage_account_key", // code plugin defines 2 regexes for this, one we want to use but the other is a bit loose, we will use just the one we want and define it in this package
}

func NewSanitizer(opts SanitizerOptions) *Sanitizer {
	s := &Sanitizer{
		excludeFields: helpers.SliceToLookup(opts.ExcludeFields),
	}

	builtInExcludeFields := opts.ExcludeFields
	codePluginMatchers := secrets.Matchers()

	for _, sm := range codePluginMatchers {
		if !slices.Contains(builtInExcludeFields, sm.Type()) {
			builtInExcludeFields = append(builtInExcludeFields, sm.Type())
		}
	}

	// first convert exclude fields to regex regexes to exclude the fields from both JSON and YAML
	for _, f := range opts.ExcludeFields {
		re, err := regexp.Compile(getExcludeFromJsonRegex(f))
		if err != nil {
			slog.Warn("Invalid regex pattern", slog.String("pattern", f), "error", err)
			continue
		}
		s.fieldPatternRegexes = append(s.fieldPatternRegexes, re)

	}

	// add in custom regexes
	for _, p := range opts.ExcludePatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			slog.Warn("Invalid regex pattern", slog.String("pattern", p), "error", err)
			continue
		}
		s.regexes = append(s.regexes, re)
	}

	if opts.ImportCodeMatchers {
		for _, sm := range codePluginMatchers {
			// basic_auth: matches URLs with a specific scheme (like http, https, ftp, etc.), followed by a user and password before an @
			if !slices.Contains(codePluginExcludedRegex, sm.Type()) {
				s.regexes = append(s.regexes, sm.DenyList()...)
			}
		}
	}
	return s
}

func (s *Sanitizer) FieldExcluded(v string) bool {
	_, excluded := s.excludeFields[v]
	return excluded
}

func (s *Sanitizer) SanitizeString(v string) string {
	type replacement struct {
		start int
		end   int
	}
	var replacements []*replacement
	// first field replacements - for this we will replace _just the first capture group_
	for _, re := range s.fieldPatternRegexes {
		for _, m := range re.FindAllStringSubmatchIndex(v, -1) {
			var startOffset, endOffset int
			if len(m) > 2 {
				// If the regexp in the secret matcher has a match group, then use it
				// as the "secret" from the string. For example "user:(secret)".
				startOffset = m[2]
				endOffset = m[3]
			} else {
				// If the regexp has no match group, then use the full match as the secret.
				// e.g. "tok-[a-z]+"
				startOffset = m[0]
				endOffset = m[1]
			}
			replacements = append(replacements, &replacement{
				start: startOffset,
				end:   endOffset,
			})
		}
	}
	// now full regex replacements - replace the full match
	for _, re := range s.regexes {
		for _, m := range re.FindAllStringSubmatchIndex(v, -1) {
			// use the full match as the secret.
			startOffset := m[0]
			endOffset := m[1]

			replacements = append(replacements, &replacement{
				start: startOffset,
				end:   endOffset,
			})
		}
	}

	// now order replacements and remove overlaps
	sort.Slice(replacements, func(i, j int) bool {
		return replacements[i].start < replacements[j].start
	})
	var newReplacements []*replacement
	var lastReplacement *replacement
	for _, r := range replacements {
		if lastReplacement != nil && r.start < lastReplacement.end {
			a := v[r.start:r.end]
			b := v[lastReplacement.start:lastReplacement.end]
			slog.Debug("Overlapping replacements", "a", a, "b", b)
			// expand previous replacement
			lastReplacement.end = r.end
			continue
		}
		newReplacements = append(newReplacements, r)
		lastReplacement = r
	}

	// now apply replacements in reverse order so the indexes remain valid
	for i := len(newReplacements) - 1; i >= 0; i-- {
		r := newReplacements[i]
		v = v[:r.start] + RedactedStr + v[r.end:]
	}
	return v
}

// Sanitize takes any value and returns a sanitized version of the value.
// If the value is a string, then it is sanitized.
// Otherwise the value is marshaled to JSON and then sanitized.
// Attempt to marshal back to original type but if this fails, return the json
func (s *Sanitizer) Sanitize(v any) any {
	valStr, isString := v.(string)

	if !isString {
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return RedactedStr
		}
		valStr = string(jsonBytes)
	}

	sanitizedString := s.SanitizeString(valStr)

	if sanitizedString == valStr {
		return v
	}
	if isString {
		return sanitizedString
	}

	// TODO slice, other types
	// maybe switch on typoe and use SanitizeStruct?
	var res = new(map[string]any)
	err := json.Unmarshal([]byte(sanitizedString), res)
	if err != nil {
		return sanitizedString
	}
	return res
}

func SanitizeStruct[T any](s *Sanitizer, v T) (T, error) {
	var empty T

	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return empty, err
	}
	valStr := string(jsonBytes)
	sanitizedString := s.SanitizeString(valStr)

	if sanitizedString == valStr {
		return v, err
	}

	err = json.Unmarshal([]byte(sanitizedString), &empty)
	return empty, err
}

func (s *Sanitizer) SanitizeKeyValue(k string, v any) any {
	if s.FieldExcluded(k) {
		return RedactedStr
	}
	return s.Sanitize(v)
}

func (s *Sanitizer) SanitizeFile(file string) error {
	// Read the contents of the file.
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	// Sanitize the string content of the file.
	sanitizedData := s.SanitizeString(string(data))

	// Write the sanitized data back to the file.
	err = os.WriteFile(file, []byte(sanitizedData), 0600)
	if err != nil {
		return err
	}

	return nil
}

// func getExcludeFromYamlRegex(fieldName string) string {
// 	return fmt.Sprintf(`(?i)%s:\s*([^\n]+)`, fieldName)
// }

// func getExcludeFromEquals(fieldName string) string {
// 	return fmt.Sprintf(`(?i)%s\s*=\s*(?:\033\[[^m]*m)*([^\033\n]+)`, fieldName)
// }

func getExcludeFromJsonRegex(fieldName string) string {
	return fmt.Sprintf(`(?i)"%s"\s*:\s*"((?:\\"|[^"])+)"`, fieldName)
}
