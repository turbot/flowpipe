package types

// This type is used by the API to return a list of integrations.
type ListIntegrationResponse struct {
	Items     []FpIntegration `json:"items"`
	NextToken *string         `json:"next_token,omitempty"`
}

type FpIntegration struct {
	Name            string  `json:"name"`
	Type            string  `json:"type"`
	Description     *string `json:"description,omitempty"`
	Title           *string `json:"title,omitempty"`
	Documentation   *string `json:"documentation,omitempty"`
	FileName        string  `json:"file_name,omitempty"`
	StartLineNumber int     `json:"start_line_number,omitempty"`
	EndLineNumber   int     `json:"end_line_number,omitempty"`

	RootMod string `json:"root_mod"`
}
