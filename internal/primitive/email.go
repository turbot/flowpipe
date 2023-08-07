package primitive

import (
	"context"
	"fmt"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"

	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/pipeparser/schema"
)

type Email struct {
	Input types.Input
}

func (h *Email) ValidateInput(ctx context.Context, i types.Input) error {

	// Validate sender's information
	if i[schema.AttributeTypeFrom] == nil {
		return fperr.BadRequestWithMessage("Email input must define from")
	}
	if i[schema.AttributeTypeSenderCredential] == nil {
		return fperr.BadRequestWithMessage("Email input must define sender_credential")
	}
	if i[schema.AttributeTypeHost] == nil {
		return fperr.BadRequestWithMessage("Email input must define a SMTP host")
	}
	if i[schema.AttributeTypePort] == nil {
		return fperr.BadRequestWithMessage("Email input must define a port")
	}
	if i[schema.AttributeTypeSenderName] != nil {
		if _, ok := i[schema.AttributeTypeSenderName].(string); !ok {
			return fperr.BadRequestWithMessage("Email attribute 'sender_name' must be a string")
		}
	}

	// Validate the recipients
	if i[schema.AttributeTypeTo] == nil {
		return fperr.BadRequestWithMessage("Email input must define to")
	}
	if _, ok := i[schema.AttributeTypeTo].([]string); !ok {
		return fperr.BadRequestWithMessage("Email attribute 'to' must be an array")
	}
	if len(i[schema.AttributeTypeTo].([]string)) == 0 {
		return fperr.BadRequestWithMessage("Recipients must not be empty")
	}

	// Validate the Cc recipients
	if i[schema.AttributeTypeCc] != nil {
		if _, ok := i[schema.AttributeTypeCc].([]string); !ok {
			return fperr.BadRequestWithMessage("Email attribute 'cc' must be an array")
		}
	}

	// Validate the Bcc recipients
	if i[schema.AttributeTypeBcc] != nil {
		if _, ok := i[schema.AttributeTypeBcc].([]string); !ok {
			return fperr.BadRequestWithMessage("Email attribute 'bcc' must be an array")
		}
	}

	// Validate the email body
	if i[schema.AttributeTypeBody] != nil {
		if _, ok := i[schema.AttributeTypeBody].(string); !ok {
			return fperr.BadRequestWithMessage("Email attribute 'body' must be a string")
		}
	}

	if i[schema.AttributeTypeContentType] != nil {
		if _, ok := i[schema.AttributeTypeContentType].(string); !ok {
			return fperr.BadRequestWithMessage("Email attribute 'content_type' must be a string")
		}
	}

	// validate the subject
	if i[schema.AttributeTypeSubject] != nil {
		if _, ok := i[schema.AttributeTypeSubject].(string); !ok {
			return fperr.BadRequestWithMessage("Email attribute 'subject' must be a string")
		}
	}

	return nil
}

func (h *Email) Run(ctx context.Context, input types.Input) (*types.Output, error) {
	// Validate the inputs
	if err := h.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	// Read sender credential
	senderEmail := input[schema.AttributeTypeFrom].(string)
	senderCredential := input[schema.AttributeTypeSenderCredential].(string)
	host := input[schema.AttributeTypeHost].(string)
	port := input[schema.AttributeTypePort].(string)
	auth := smtp.PlainAuth("", senderEmail, senderCredential, host)

	// Get the inputs
	recipients := input[schema.AttributeTypeTo].([]string)

	var body string
	if input[schema.AttributeTypeBody] != nil {
		body = input[schema.AttributeTypeBody].(string)
	}

	var senderName string
	if input[schema.AttributeTypeSenderName] != nil {
		senderName = input[schema.AttributeTypeSenderName].(string)
	}

	from := mail.Address{
		Name:    senderName,
		Address: senderEmail,
	}

	// Construct the header
	header := make(map[string]string)
	header["From"] = from.String()
	header["To"] = strings.Join(recipients, ", ")

	if input[schema.AttributeTypeCc] != nil && len(input[schema.AttributeTypeCc].([]string)) > 0 {
		header["Cc"] = strings.Join(input[schema.AttributeTypeCc].([]string), ", ")
	}

	if input[schema.AttributeTypeBcc] != nil && len(input[schema.AttributeTypeBcc].([]string)) > 0 {
		header["Bcc"] = strings.Join(input[schema.AttributeTypeBcc].([]string), ", ")
	}

	if input[schema.AttributeTypeSubject] != nil && len(input[schema.AttributeTypeSubject].(string)) > 0 {
		header["Subject"] = input[schema.AttributeTypeSubject].(string)
	}

	if input[schema.AttributeTypeContentType] != nil && len(input[schema.AttributeTypeContentType].(string)) > 0 {
		header["Content-Type"] = input[schema.AttributeTypeContentType].(string)
	}

	// Build the full email message
	var message string
	for key, value := range header {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	message += "\r\n" + body

	// Construct the output
	output := types.Output{
		Data: map[string]interface{}{},
	}

	start := time.Now().UTC()
	err := smtp.SendMail(host+":"+port, auth, senderEmail, recipients, []byte(message))
	finish := time.Now().UTC()
	if err != nil {
		if _, ok := err.(*textproto.Error); !ok {
			return nil, err
		}

		// Capture all 400+ errors related to negative completion in the output
		// Refer https://en.wikipedia.org/wiki/List_of_SMTP_server_return_codes for all available error codes
		smtpErr := err.(*textproto.Error)
		if smtpErr.Code >= 400 {
			output.Errors = []types.StepError{
				types.StepError{
					Message:   smtpErr.Msg,
					ErrorCode: smtpErr.Code,
				},
			}
		}
	}

	output.Data[schema.AttributeTypeStartedAt] = start
	output.Data[schema.AttributeTypeFinishedAt] = finish

	return &output, nil
}
