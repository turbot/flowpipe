package sanitize

import (
	"testing"
)

//
//func TestSanitizer_SanitizeKeyValue(t *testing.T) {
//
//	tests := []struct {
//		name string
//		opts SanitizerOptions
//		args []any
//		want []any
//	}{
//		{
//			name: "simple regex",
//			opts: SanitizerOptions{
//				ExcludePatterns: []string{"password:(\\s*\\S+)"},
//			},
//			args: []any{
//				"key1", "password:foo",
//			},
//			want: []any{
//				"key1", "password:" + redactedStr,
//			},
//		},
//		{
//			name: "multiple capture groups",
//			opts: SanitizerOptions{
//				ExcludePatterns: []string{"password:(\\s*\\S+) token:(\\s*\\S+)"},
//			},
//			args: []any{
//				"key1", "password:foo token:bar",
//			},
//			want: []any{
//				"key1", fmt.Sprintf("password:%s token:%s", redactedStr, redactedStr),
//			},
//		},
//		{
//			name: "no capture groups",
//			opts: SanitizerOptions{
//				ExcludePatterns: []string{"password:\\s*\\S+"},
//			},
//			args: []any{
//				"key1", "password:foo",
//			},
//			want: []any{
//				"key1", "password:foo",
//			},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := NewSanitizer(tt.opts)
//
//			if got := s.SanitizeKeyValues(tt.args...); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("SanitizeKeyValues() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

func TestSanitizer_SanitizeString(t *testing.T) {

	tests := []struct {
		name  string
		opts  SanitizerOptions
		input string
		want  string
	}{
		{
			name: "simple regex",
			opts: SanitizerOptions{
				ExcludeFields: []string{"password"},
			},
			input: `{"password":"foo"}`,
			want:  `{"password":"` + redactedStr + `"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSanitizer(tt.opts)
			if got := s.SanitizeString(tt.input); got != tt.want {
				t.Errorf("SanitizeString() = %v, want %v", got, tt.want)
			}
		})
	}
}
