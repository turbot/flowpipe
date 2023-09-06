package primitive

import (
	"context"
	"fmt"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"

	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/perr"
	"github.com/turbot/flowpipe/pipeparser/schema"
)

type Email struct {
	Input modconfig.Input
}

func (h *Email) ValidateInput(ctx context.Context, i modconfig.Input) error {

	// Validate sender's information
	if i[schema.AttributeTypeFrom] == nil {
		return perr.BadRequestWithMessage("Email input must define from")
	}
	if i[schema.AttributeTypeSenderCredential] == nil {
		return perr.BadRequestWithMessage("Email input must define sender_credential")
	}
	if i[schema.AttributeTypeHost] == nil {
		return perr.BadRequestWithMessage("Email input must define a SMTP host")
	}
	if i[schema.AttributeTypePort] == nil {
		return perr.BadRequestWithMessage("Email input must define a port")
	}
	if i[schema.AttributeTypeSenderName] != nil {
		if _, ok := i[schema.AttributeTypeSenderName].(string); !ok {
			return perr.BadRequestWithMessage("Email attribute 'sender_name' must be a string")
		}
	}

	// Validate the recipients
	if i[schema.AttributeTypeTo] == nil {
		return perr.BadRequestWithMessage("Email input must define to")
	}
	if _, ok := i[schema.AttributeTypeTo].([]string); !ok {
		// The given input is a string slice, but the step input stores it as an interface slice during the JSON unmarshalling?
		// Check if the input is an interface slice, and the elements are strings
		if data, ok := i[schema.AttributeTypeTo].([]interface{}); ok {
			for _, v := range data {
				if _, ok := v.(string); !ok {
					return perr.BadRequestWithMessage("Email attribute 'to' must have elements of type string")
				}
			}
			return nil
		}

		return perr.BadRequestWithMessage("Email attribute 'to' must be an array")
	}

	var recipients []string
	if _, ok := i[schema.AttributeTypeTo].([]string); ok {
		recipients = i[schema.AttributeTypeTo].([]string)
	}

	if _, ok := i[schema.AttributeTypeTo].([]interface{}); ok {
		for _, v := range i[schema.AttributeTypeTo].([]interface{}) {
			recipients = append(recipients, v.(string))
		}
	}

	if len(recipients) == 0 {
		return perr.BadRequestWithMessage("Recipients must not be empty")
	}

	// Validate the Cc recipients
	if i[schema.AttributeTypeCc] != nil {
		if _, ok := i[schema.AttributeTypeCc].([]string); !ok {
			// The given input is a string slice, but the step input stores it as an interface slice during the JSON unmarshalling?
			// Check if the input is an interface slice, and the elements are strings
			if data, ok := i[schema.AttributeTypeCc].([]interface{}); ok {
				for _, v := range data {
					if _, ok := v.(string); !ok {
						return perr.BadRequestWithMessage("Email attribute 'cc' must have elements of type string")
					}
				}
				return nil
			}

			return perr.BadRequestWithMessage("Email attribute 'cc' must be an array")
		}
	}

	// Validate the Bcc recipients
	if i[schema.AttributeTypeBcc] != nil {
		if _, ok := i[schema.AttributeTypeBcc].([]string); !ok {
			// The given input is a string slice, but the step input stores it as an interface slice during the JSON unmarshalling?
			// Check if the input is an interface slice, and the elements are strings
			if data, ok := i[schema.AttributeTypeBcc].([]interface{}); ok {
				for _, v := range data {
					if _, ok := v.(string); !ok {
						return perr.BadRequestWithMessage("Email attribute 'bcc' must have elements of type string")
					}
				}
				return nil
			}

			return perr.BadRequestWithMessage("Email attribute 'bcc' must be an array")
		}
	}

	// Validate the email body
	if i[schema.AttributeTypeBody] != nil {
		if _, ok := i[schema.AttributeTypeBody].(string); !ok {
			return perr.BadRequestWithMessage("Email attribute 'body' must be a string")
		}
	}

	if i[schema.AttributeTypeContentType] != nil {
		if _, ok := i[schema.AttributeTypeContentType].(string); !ok {
			return perr.BadRequestWithMessage("Email attribute 'content_type' must be a string")
		}
	}

	// validate the subject
	if i[schema.AttributeTypeSubject] != nil {
		if _, ok := i[schema.AttributeTypeSubject].(string); !ok {
			return perr.BadRequestWithMessage("Email attribute 'subject' must be a string")
		}
	}

	return nil
}

func (h *Email) Run(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {
	// Validate the inputs
	if err := h.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	// Read sender credential
	senderEmail := input[schema.AttributeTypeFrom].(string)
	senderCredential := input[schema.AttributeTypeSenderCredential].(string)
	host := input[schema.AttributeTypeHost].(string)
	port := input[schema.AttributeTypePort].(int64)
	auth := smtp.PlainAuth("", senderEmail, senderCredential, host)

	// Get the inputs
	var recipients []string
	if _, ok := input[schema.AttributeTypeTo].([]string); ok {
		recipients = input[schema.AttributeTypeTo].([]string)
	}

	if _, ok := input[schema.AttributeTypeTo].([]interface{}); ok {
		for _, v := range input[schema.AttributeTypeTo].([]interface{}) {
			recipients = append(recipients, v.(string))
		}
	}

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

	if input[schema.AttributeTypeSubject] != nil && len(input[schema.AttributeTypeSubject].(string)) > 0 {
		header["Subject"] = input[schema.AttributeTypeSubject].(string)
	}

	if input[schema.AttributeTypeContentType] != nil && len(input[schema.AttributeTypeContentType].(string)) > 0 {
		header["Content-Type"] = input[schema.AttributeTypeContentType].(string)
	}

	if input[schema.AttributeTypeCc] != nil {
		var cc []string

		// Check if the input is a string slice
		if _, ok := input[schema.AttributeTypeCc].([]string); ok {
			cc = input[schema.AttributeTypeCc].([]string)
		}

		// Check if the input is an interface slice, and the elements are strings
		if _, ok := input[schema.AttributeTypeCc].([]interface{}); ok {
			for _, v := range input[schema.AttributeTypeCc].([]interface{}) {
				cc = append(cc, v.(string))
			}
		}

		// if the cc is not empty, add it to the header
		if len(cc) > 0 {
			header["Cc"] = strings.Join(cc, ", ")
		}
	}

	if input[schema.AttributeTypeBcc] != nil {
		var bcc []string

		// Check if the input is a string slice
		if _, ok := input[schema.AttributeTypeBcc].([]string); ok {
			bcc = input[schema.AttributeTypeBcc].([]string)
		}

		// Check if the input is an interface slice, and the elements are strings
		if _, ok := input[schema.AttributeTypeBcc].([]interface{}); ok {
			for _, v := range input[schema.AttributeTypeBcc].([]interface{}) {
				bcc = append(bcc, v.(string))
			}
		}

		// if the cc is not empty, add it to the header
		if len(bcc) > 0 {
			header["Bcc"] = strings.Join(bcc, ", ")
		}
	}

	// Build the full email message
	var message string
	for key, value := range header {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	message += "\r\n" + body

	// Construct the output
	output := modconfig.Output{
		Data: map[string]interface{}{},
	}

	// Build the address of the SMTP server
	addr := host + ":" + fmt.Sprintf("%d", port)

	start := time.Now().UTC()
	err := smtp.SendMail(addr, auth, senderEmail, recipients, []byte(message))
	finish := time.Now().UTC()
	if err != nil {
		if _, ok := err.(*textproto.Error); !ok {
			return nil, err
		}

		// Capture all 400+ errors related to negative completion in the output
		// Refer https://en.wikipedia.org/wiki/List_of_SMTP_server_return_codes for all available error codes
		smtpErr := err.(*textproto.Error)
		if smtpErr.Code >= 400 {
			output.Errors = []modconfig.StepError{
				{
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
