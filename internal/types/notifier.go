package types

import "github.com/turbot/pipe-fittings/modconfig"

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

func FpNotifierFromModNotifier(notifier modconfig.Notifier) (*FpNotifier, error) {
	resp := &FpNotifier{
		Name: notifier.Name(),
	}

	resp.FileName = notifier.GetNotifierImpl().FileName
	resp.StartLineNumber = notifier.GetNotifierImpl().StartLineNumber
	resp.EndLineNumber = notifier.GetNotifierImpl().EndLineNumber

	return resp, nil
}
