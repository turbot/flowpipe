package sanitize

import (
	"testing"
)

func TestSanitizer_SanitizeString(t *testing.T) {
	// TODO
	//	tests := []struct {
	//		name  string
	//		opts  SanitizerOptions
	//		input string
	//		want  string
	//	}{
	//		{
	//			name: "simple regex",
	//			opts: SanitizerOptions{
	//				ExcludeFields: []string{"password"},
	//			},
	//			input: `{"password":"foo"}`,
	//			want:  `{"password":"` + redactedStr + `"}`,
	//		},
	//		{
	//			name: "pipeline list",
	//			opts: SanitizerOptions{
	//				ExcludeFields: []string{"name"},
	//			},
	//			input: `[
	//   {
	//      "name":"default_mod.pipeline.echo_one",
	//      "mod":"mod.default_mod",
	//      "steps":[
	//         {
	//            "name":"echo_one",
	//            "step_type":"transform",
	//            "pipeline_name":"REDACTED",
	//            "value":"Hello World"
	//         },
	//         {
	//            "name":"child_pipeline",
	//            "step_type":"pipeline",
	//            "pipeline_name":"REDACTED",
	//            "args":null
	//         }
	//      ],
	//      "outputs":[
	//         {
	//            "name":"echo_one_output"
	//         }
	//      ]
	//   }
	//]`,
	//			want: `{"password":"` + redactedStr + `"}`,
	//		},
	//	}
	//	for _, tt := range tests {
	//		t.Run(tt.name, func(t *testing.T) {
	//			s := NewSanitizer(tt.opts)
	//			if got := s.SanitizeString(tt.input); got != tt.want {
	//				t.Errorf("SanitizeString() = %v, want %v", got, tt.want)
	//			}
	//		})
	//	}
}
