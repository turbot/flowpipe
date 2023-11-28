package sanitize

import (
	"reflect"
	"slices"
	"unicode"
)

const redactedStr = "<redacted>"

type SanitizerOptions struct {
	// ExcludeFields is a list of fields to exclude from sanitization
	ExcludeFields []string
	// ExcludePatterns is a list of patterns to exclude from sanitization
	ExcludePatterns []string
}

type Sanitizer struct {
	opts SanitizerOptions
}

func NewSanitizer(opts SanitizerOptions) *Sanitizer {
	return &Sanitizer{
		opts: opts,
	}
}

func (s *Sanitizer) SanitizeKeyValue(keysAndValues ...any) []any {

	// TODO better to just let the logging library do this?
	//if len(keysAndValues)%2 != 0 {
	//	// empty the whole thing if the keys and values are not in pairs
	//	return nil
	//}

	sanitizeKeyAndValues := make([]any, len(keysAndValues))

	for i := 0; i < len(keysAndValues); i += 2 {
		k := keysAndValues[i]
		// write key
		sanitizeKeyAndValues[i] = k

		// check for missing final value
		if i+1 >= len(keysAndValues) {
			break
		}

		v := keysAndValues[i+1]

		if reflect.ValueOf(v).Kind() == reflect.Struct {
			// if value is a struct, sanitize it
			val := s.SanitizeStruct(v)
			sanitizeKeyAndValues[i+1] = val
		} else {
			s.SanitizeSimpleKeyValue(k, v)
		}
	}

	return sanitizeKeyAndValues
}

func (s *Sanitizer) SanitizeStruct(v any) map[string]any {
	results := make(map[string]any)

	val := reflect.ValueOf(v) // could be any underlying type

	// if its a pointer, resolve its value
	if val.Kind() == reflect.Ptr {
		val = reflect.Indirect(val)
	}

	// should double check we now have a struct (could still be anything)
	if val.Kind() != reflect.Struct {
		return results
	}

	// now we grab our values as before (note: I assume table name should come from the struct type)
	structType := val.Type()

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		fieldName := field.Name

		if unicode.IsLower(rune(fieldName[0])) {
			// private field, no need to redact
			continue
		}

		var sanitizedValue any
		if s.shouldExclude(fieldName) {
			// if the field is in the exclude list, then redact the value
			sanitizedValue = redactedStr
		} else {
			// otherwise, get the value and if it is a struct, recurse in to sanitize
			value := val.FieldByName(fieldName)
			if value.Kind() == reflect.Ptr {
				sanitizedValue = s.SanitizeStruct(reflect.Indirect(val))
			} else if value.Kind() == reflect.Struct {
				sanitizedValue = s.SanitizeStruct(value.Interface())
			} else {
				// non struct - just return the value unaltered
				sanitizedValue = val.FieldByName(fieldName).Interface()
			}
		}

		results[fieldName] = sanitizedValue
	}
	return results

}

func (s *Sanitizer) SanitizeSimpleKeyValue(k, v any) any {
	keyString, ok := k.(string)
	if ok && s.shouldExclude(keyString) {
		return redactedStr
	}
	return v
}

func (s *Sanitizer) shouldExclude(fieldName string) bool {
	return s.opts.ExcludeFields != nil && slices.Contains(s.opts.ExcludeFields, fieldName)
}
