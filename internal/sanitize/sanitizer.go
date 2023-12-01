package sanitize

import (
	"encoding/json"
	"fmt"
	"github.com/turbot/go-kit/helpers"
	"log"
	"log/slog"
	"regexp"
	"sort"
)

const redactedStr = "<redacted>"

// TODO where should this be defined
var Instance = NewSanitizer(SanitizerOptions{
	ExcludeFields: []string{
		"pipeline_execution_id",
		"pipeline_name",
		//"mod",
		//"step_type",
		" description",
		"value",
		"foo",
	},
	//ExcludePatterns: []string{
	//	"Starting",
	//},
})

type SanitizerOptions struct {
	// ExcludeFields is a list of fields to exclude from sanitization
	ExcludeFields []string
	// ExcludePatterns is a list of regexes - any capture groups are redacted
	ExcludePatterns []string
}

type Sanitizer struct {
	patterns      []*regexp.Regexp
	excludeFields map[string]struct{}
}

func NewSanitizer(opts SanitizerOptions) *Sanitizer {
	// dedupe patterns using map
	var patterns = make(map[string]struct{}, len(opts.ExcludeFields)+len(opts.ExcludePatterns))

	// first convert exclude fields to regex patterns to exclude the fields from both JSON and YAML
	for _, f := range opts.ExcludeFields {
		patterns[getExcludeFromJsonRegex(f)] = struct{}{}
		patterns[getExcludeFromYamlRegex(f)] = struct{}{}
		patterns[getExcludeFromEquals(f)] = struct{}{}
	}

	// add in custom patterns
	for _, p := range opts.ExcludePatterns {
		patterns[p] = struct{}{}
	}

	s := &Sanitizer{
		excludeFields: helpers.SliceToLookup(opts.ExcludeFields),
	}

	// now convert all patterns into regexes
	for p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			slog.Warn("Invalid regex pattern", slog.String("pattern", p), "error", err)
			continue
		}
		s.patterns = append(s.patterns, re)
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
	for _, re := range s.patterns {
		matchGroups := re.FindAllStringSubmatchIndex(v, -1)
		for _, m := range matchGroups {
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
			log.Printf("Overlapping replacements: %s and %s", a, b)
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
		v = v[:r.start] + redactedStr + v[r.end:]
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
			return redactedStr
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
		return redactedStr
	}
	return s.Sanitize(v)
}

func (s *Sanitizer) SanitizeFile(string) {

}

func getExcludeFromYamlRegex(fieldName string) string {
	return fmt.Sprintf(`%s:\s*([^\n]+)`, fieldName)
}

func getExcludeFromEquals(fieldName string) string {
	return fmt.Sprintf(`%s\s*=\s*(?:\033\[[^m]*m)*([^\033\n]+)`, fieldName)
}

func getExcludeFromJsonRegex(fieldName string) string {
	return fmt.Sprintf(`"%s"\s*:\s*"([^"]+)"`, fieldName)
}
