package util

import (
	"context"
	"fmt"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"

	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
)

func RunSendEmail(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {

	// Read sender credential
	senderEmail := input[schema.AttributeTypeFrom].(string)
	senderCredential := input[schema.AttributeTypeSenderCredential].(string)
	host := input[schema.AttributeTypeHost].(string)
	auth := smtp.PlainAuth("", senderEmail, senderCredential, host)

	// Convert port into integer
	var portInt int64
	if port, ok := input[schema.AttributeTypePort].(float64); ok {
		portInt = int64(port)
	}
	if port, ok := input[schema.AttributeTypePort].(int64); ok {
		portInt = port
	}

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
	addr := host + ":" + fmt.Sprintf("%d", portInt)

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
