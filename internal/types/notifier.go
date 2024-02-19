package types

type ListNotifierResponse struct {
	Items     []FpNotifier `json:"items"`
	NextToken *string      `json:"next_token,omitempty"`
}

type FpNotifier struct {
	Name            string  `json:"name"`
	Description     *string `json:"description,omitempty"`
	Title           *string `json:"title,omitempty"`
	Documentation   *string `json:"documentation,omitempty"`
	FileName        string  `json:"file_name,omitempty"`
	StartLineNumber int     `json:"start_line_number,omitempty"`
	EndLineNumber   int     `json:"end_line_number,omitempty"`
}
