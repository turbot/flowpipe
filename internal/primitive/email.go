package primitive

import (
	"context"
	"fmt"
	"net/mail"
	"net/smtp"
	"os"
	"strings"

	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/pipeparser/schema"
)

type Email struct {
	Input   types.Input
	Setting string
}

func (h *Email) ValidateInput(ctx context.Context, i types.Input) error {
	if i[schema.AttributeTypeTo] == nil {
		return fperr.BadRequestWithMessage("Email input must define a recipients")
	}

	if len(i[schema.AttributeTypeTo].([]string)) == 0 {
		return fperr.BadRequestWithMessage("Recipients must not be empty")
	}

	// A body is required as we will be sending the output of the previous step as the body of the email, and in that case, the body will never be empty
	if i[schema.AttributeTypeBody] == nil {
		return fperr.BadRequestWithMessage("Email input must define a body")
	}

	return nil
}

func (h *Email) Run(ctx context.Context, input types.Input) (*types.StepOutput, error) {
	// Validate the inputs
	if err := h.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	// Read sender credential
	var senderEmail, senderCredential, host, port string
	if h.Setting == "mailhog" { // For testing purposes
		senderEmail = "sender@example.com"
		senderCredential = "" // Empty for Mailhog
		host = "localhost"
		port = "1025"
	} else {
		senderEmail = os.Getenv("SMTP_SENDER_EMAIL") // Check for other options. What email should we use?
		senderCredential = os.Getenv("SMTP_SENDER_CREDENTIAL")
		host = "smtp.gmail.com" // Should be dynamic? Or should be a fixed value since the sender will be always flowpipe?
		port = "587"
	}
	auth := smtp.PlainAuth("", senderEmail, senderCredential, host)

	// Get the inputs
	var ccRecipients, bccRecipients []string
	recipients := input[schema.AttributeTypeTo].([]string)
	if input[schema.AttributeTypeCc] != nil {
		ccRecipients = input[schema.AttributeTypeCc].([]string)
	}
	if input[schema.AttributeTypeBcc] != nil {
		bccRecipients = input[schema.AttributeTypeBcc].([]string)
	}

	var body string
	if input[schema.AttributeTypeBody] != nil {
		body = input[schema.AttributeTypeBody].(string)
	}

	from := mail.Address{
		Name:    "Flowpipe",
		Address: senderEmail,
	}

	// Construct the MIME header
	header := make(map[string]string)
	header["From"] = from.String()
	header["Subject"] = "Flowpipe pipeline completion notification" // What should be the subject? Should we assume that it would be in the input?
	header["To"] = strings.Join(recipients, ", ")

	if len(ccRecipients) > 0 {
		header["Cc"] = strings.Join(ccRecipients, ", ")
	}
	if len(bccRecipients) > 0 {
		header["Bcc"] = strings.Join(bccRecipients, ", ")
	}

	// Build the full email message
	var message string
	for key, value := range header {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	message += "\r\n" + body

	err := smtp.SendMail(host+":"+port, auth, senderEmail, recipients, []byte(message))
	if err != nil {
		return nil, err
	}

	return nil, nil
}
