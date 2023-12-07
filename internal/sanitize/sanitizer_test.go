package sanitize

import (
	"testing"
)

func TestSanitizer_SanitizeString(t *testing.T) {
	tests := []struct {
		name  string
		opts  SanitizerOptions
		input string
		want  string
	}{
		{
			name: "field value replacement",
			opts: SanitizerOptions{
				ExcludeFields: []string{"password"},
			},
			input: `{"password":"foo"}`,
			want:  `{"password":"` + RedactedStr + `"}`,
		},
		{
			name: "full replacement",
			opts: SanitizerOptions{
				ExcludePatterns: []string{"mypass([0-9]*)"},
			},
			input: `{"password":"mypass12345"}`,
			want:  `{"password":"` + RedactedStr + `"}`,
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
