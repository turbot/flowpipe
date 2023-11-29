package sanitize

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"
)

const redactedStr = "<redacted>"

type SanitizerOptions struct {
	// ExcludeFields is a list of fields to exclude from sanitization
	ExcludeFields []string
	// ExcludePatterns is a list of regexes - any capture groups are redacted
	ExcludePatterns []string
}

type Sanitizer struct {
	patterns []*regexp.Regexp
}

func NewSanitizer(opts SanitizerOptions) *Sanitizer {
	// dedupe patterns using map
	var patterns = make(map[string]struct{}, len(opts.ExcludeFields)+len(opts.ExcludePatterns))

	// first convert exclude fields to regex patterns to exclude the fields from both JSON and YAML
	for _, f := range opts.ExcludeFields {
		excludeFromJson := getExcludeFromJsonRegex(f)
		patterns[excludeFromJson] = struct{}{}

		excludeFromYaml := getExcludeFromYamlRegex(f)
		patterns[excludeFromYaml] = struct{}{}
	}

	// add in custom patterns
	for _, p := range opts.ExcludePatterns {
		patterns[p] = struct{}{}
	}

	s := &Sanitizer{}

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

func getExcludeFromYamlRegex(fieldName string) string {
	return fmt.Sprintf(`%s:\s*([^\n]+)`, fieldName)

}

func getExcludeFromJsonRegex(fieldName string) string {
	return fmt.Sprintf(`"%s"\s*:\s*"([^"]+)"`, fieldName)
}

//// TODO KAI MAPS/SLICES
//
//func (s *Sanitizer) SanitizeKeyValues(keysAndValues ...any) []any {
//
//	// TODO better to just let the logging library do this?
//	//if len(keysAndValues)%2 != 0 {
//	//	// empty the whole thing if the keys and values are not in pairs
//	//	return nil
//	//}
//
//	sanitizeKeyAndValues := make([]any, len(keysAndValues))
//
//	for i := 0; i < len(keysAndValues); i += 2 {
//		k := keysAndValues[i]
//		// write key
//		sanitizeKeyAndValues[i] = k
//
//		// check for missing final value
//		if i+1 >= len(keysAndValues) {
//			break
//		}
//		v := keysAndValues[i+1]
//
//		sanitizeKeyAndValues[i+1] = s.SanitizeKeyValue(k, v)
//	}
//
//	return sanitizeKeyAndValues
//}
//
//func (s *Sanitizer) SanitizeKeyValue(k, v any) any {
//	// first check if the key is in the exclude list
//	keyString, ok := k.(string)
//	if ok && s.shouldExclude(keyString) {
//		return redactedStr
//	}
//
//	var sanitizedValue = v
//	switch reflect.ValueOf(v).Kind() {
//	case reflect.String:
//		sanitizedValue = s.SanitizeString(v.(string))
//	case reflect.Struct:
//		sanitizedValue = s.SanitizeStruct(v)
//	case reflect.Map:
//		sanitizedValue = s.SanitizeMap(v)
//	case reflect.Slice:
//		sanitizedValue = s.SanitizeSlice(v)
//	}
//
//	return sanitizedValue
//}
//
//func (s *Sanitizer) SanitizeStruct(v any) map[string]any {
//	results := make(map[string]any)
//
//	val := reflect.ValueOf(v) // could be any underlying type
//
//	// if its a pointer, resolve its value
//	if val.Kind() == reflect.Ptr {
//		val = reflect.Indirect(val)
//	}
//
//	// should double check we now have a struct (could still be anything)
//	if val.Kind() != reflect.Struct {
//		return results
//	}
//
//	// now we grab our values as before (note: I assume table name should come from the struct type)
//	structType := val.Type()
//
//	for i := 0; i < structType.NumField(); i++ {
//		field := structType.Field(i)
//
//		fieldName := field.Name
//
//		if unicode.IsLower(rune(fieldName[0])) {
//			// private field, no need to redact
//			// TODO check?
//			continue
//		}
//
//		var sanitizedValue any
//		if s.shouldExclude(fieldName) {
//			// if the field is in the exclude list, then redact the value
//			sanitizedValue = redactedStr
//		} else {
//			// otherwise, get the value and if it is a struct, recurse in to sanitize
//			value := val.FieldByName(fieldName)
//			if value.Kind() == reflect.Ptr {
//				sanitizedValue = s.SanitizeStruct(reflect.Indirect(val))
//			} else if value.Kind() == reflect.Struct {
//				sanitizedValue = s.SanitizeStruct(value.Interface())
//			} else {
//				// non struct - just return the value unaltered
//				sanitizedValue = val.FieldByName(fieldName).Interface()
//			}
//		}
//
//		results[fieldName] = sanitizedValue
//	}
//	return results
//
//}
//
//func (s *Sanitizer) SanitizeMap(v any) map[string]any {
//	results := make(map[string]any)
//	value := reflect.ValueOf(v)
//
//	// verify the input value is a map
//	if value.Kind() != reflect.Map {
//		panic("Input is not a map")
//	}
//
//	// Iterate over the map keys
//	for _, key := range value.MapKeys() {
//		// we only handle string keys
//		if key.Kind() != reflect.String {
//			return nil
//		}
//		// retrieve the value
//		mapValue := value.MapIndex(key)
//		// sanitize the value ands write back
//		results[key.String()] = s.SanitizeKeyValue(key, mapValue)
//	}
//
//	return results
//}
//
//func (s *Sanitizer) SanitizeSlice(v any) map[string]any {
//	results := make(map[string]any)
//
//	val := reflect.ValueOf(v) // could be any underlying type
//
//	// if its a pointer, resolve its value
//	if val.Kind() == reflect.Ptr {
//		val = reflect.Indirect(val)
//	}
//
//	// should double check we now have a struct (could still be anything)
//	if val.Kind() != reflect.Struct {
//		return results
//	}
//
//	// now we grab our values as before (note: I assume table name should come from the struct type)
//	structType := val.Type()
//
//	for i := 0; i < structType.NumField(); i++ {
//		field := structType.Field(i)
//
//		fieldName := field.Name
//
//		if unicode.IsLower(rune(fieldName[0])) {
//			// private field, no need to redact
//			// TODO check?
//			continue
//		}
//
//		var sanitizedValue any
//		if s.shouldExclude(fieldName) {
//			// if the field is in the exclude list, then redact the value
//			sanitizedValue = redactedStr
//		} else {
//			// otherwise, get the value and if it is a struct, recurse in to sanitize
//			value := val.FieldByName(fieldName)
//			if value.Kind() == reflect.Ptr {
//				sanitizedValue = s.SanitizeStruct(reflect.Indirect(val))
//			} else if value.Kind() == reflect.Struct {
//				sanitizedValue = s.SanitizeStruct(value.Interface())
//			} else {
//				// non struct - just return the value unaltered
//				sanitizedValue = val.FieldByName(fieldName).Interface()
//			}
//		}
//
//		results[fieldName] = sanitizedValue
//	}
//	return results
//
//}
//
//func (s *Sanitizer) shouldExclude(fieldName string) bool {
//	return s.opts.ExcludeFields != nil && slices.Contains(s.opts.ExcludeFields, fieldName)
//}

func (s *Sanitizer) SanitizeString(v string) string {
	for _, re := range s.patterns {
		v = re.ReplaceAllStringFunc(v, func(s string) string {
			matched := re.FindStringSubmatch(s)
			for i := 1; i < len(matched); i++ {
				v = strings.ReplaceAll(v, matched[i], redactedStr)
			}
			return v
		})
	}

	return v
}
