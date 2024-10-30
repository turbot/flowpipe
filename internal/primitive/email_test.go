package primitive

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

var mailhogCmd *exec.Cmd

func startMailHog() {
	// Start MailHog server as a separate process
	slog.Debug("Starting MailHog SMTP server")
	mailhogCmd = exec.Command("MailHog")
	if err := mailhogCmd.Start(); err != nil {
		slog.Error("Failed to start MailHog", "error", err.Error())
	}
	slog.Debug("MailHog SMTP server started")
}

func stopMailHog() {
	// Stop MailHog server process
	slog.Debug("Stopping MailHog SMTP server")
	if mailhogCmd.Process != nil {
		if err := mailhogCmd.Process.Kill(); err != nil {
			slog.Error("Failed to stop MailHog", "error", err.Error())
		}
	}
	slog.Debug("MailHog SMTP server stopped")
}

func TestSendEmail(t *testing.T) {
	assert := assert.New(t)
	hr := Email{}

	// Use a dummy SMTP server for testing (e.g., MailHog)
	input := resources.Input(map[string]interface{}{
		schema.AttributeTypeSenderName:   "TestSendEmail",
		schema.AttributeTypeFrom:         "test.send.email@example.com",
		schema.AttributeTypeSmtpPassword: "",
		schema.AttributeTypeSmtpUsername: "test.send.email@example.com",
		schema.AttributeTypeHost:         "localhost",
		schema.AttributeTypePort:         int64(1025),
		schema.AttributeTypeTo:           []string{"recipient1@example.com", "recipient2@example.com"},
		schema.AttributeTypeSubject:      "Flowpipe mail test",
		schema.AttributeTypeBody:         "This is a test email sent from Golang.",
	})

	_, err := hr.Run(context.Background(), input)
	// No errors
	assert.Nil(err)

	// Get the captured email data from the SMTP server (e.g., MailHog)
	capturedEmails, err := captureEmailsFromSMTP(input[schema.AttributeTypeFrom].(string))
	if err != nil {
		assert.Fail("error listing captured emails from Mailhog: ", err.Error())
	}

	// Check if the captured email is as expected
	if len(capturedEmails) != 1 {
		assert.Fail("Expected 1 email, but got %d", len(capturedEmails))
	}
	capturedEmail := capturedEmails[0]

	// Validate sender's information
	assert.Contains(capturedEmail.Content.Headers.From[0], "test.send.email@example.com")

	// Validate recipients
	assert.Equal([]string{"recipient1@example.com, recipient2@example.com"}, capturedEmail.Content.Headers.To)

	// Validate email body
	assert.Contains(capturedEmail.Content.Body, "This is a test email sent from Golang.")
}

func TestSendEmailWithCc(t *testing.T) {
	assert := assert.New(t)
	hr := Email{}

	// Use a dummy SMTP server for testing (e.g., MailHog)
	input := resources.Input(map[string]interface{}{
		schema.AttributeTypeSenderName:   "TestSendEmailWithCc",
		schema.AttributeTypeFrom:         "test.send.email.with.cc@example.com",
		schema.AttributeTypeSmtpUsername: "test.send.email.with.cc@example.com",
		schema.AttributeTypeSmtpPassword: "",
		schema.AttributeTypeHost:         "localhost",
		schema.AttributeTypePort:         int64(1025),
		schema.AttributeTypeTo:           []string{"recipient1@example.com", "recipient2@example.com"},
		schema.AttributeTypeCc:           []string{"ccrecipient@example.com"},
		schema.AttributeTypeBody:         "This is a test email sent from Golang with Cc.",
	})

	_, err := hr.Run(context.Background(), input)
	// No errors
	assert.Nil(err)

	// Get the captured email data from the SMTP server (e.g., MailHog)
	capturedEmails, err := captureEmailsFromSMTP(input[schema.AttributeTypeFrom].(string))
	if err != nil {
		assert.Fail("error listing captured emails from Mailhog: ", err.Error())
	}

	// Check if the captured email is as expected
	if len(capturedEmails) != 1 {
		assert.Fail("Expected 1 email, but got %d", len(capturedEmails))
	}
	capturedEmail := capturedEmails[0]

	// Validate sender's information
	assert.Contains(capturedEmail.Content.Headers.From[0], "test.send.email.with.cc@example.com")

	// Validate recipients
	assert.Equal([]string{"recipient1@example.com, recipient2@example.com"}, capturedEmail.Content.Headers.To)

	// Validate Cced recipients
	assert.Equal([]string{"ccrecipient@example.com"}, capturedEmail.Content.Headers.Cc)

	// Validate email body
	assert.Contains(capturedEmail.Content.Body, "This is a test email sent from Golang with Cc.")
}

func TestSendEmailWithBcc(t *testing.T) {
	assert := assert.New(t)
	hr := Email{}

	// Use a dummy SMTP server for testing (e.g., MailHog)
	input := resources.Input(map[string]interface{}{
		schema.AttributeTypeSenderName:   "TestSendEmailWithBcc",
		schema.AttributeTypeFrom:         "test.send.email.with.bcc@example.com",
		schema.AttributeTypeSmtpUsername: "test.send.email.with.bcc@example.com",
		schema.AttributeTypeSmtpPassword: "",
		schema.AttributeTypeHost:         "localhost",
		schema.AttributeTypePort:         int64(1025),
		schema.AttributeTypeTo:           []string{"recipient1@example.com", "recipient2@example.com"},
		schema.AttributeTypeBcc:          []string{"bccrecipient@example.com"},
		schema.AttributeTypeBody:         "This is a test email sent from Golang with Bcc.",
	})

	_, err := hr.Run(context.Background(), input)
	// No errors
	assert.Nil(err)

	// Get the captured email data from the SMTP server (e.g., MailHog)
	capturedEmails, err := captureEmailsFromSMTP(input[schema.AttributeTypeFrom].(string))
	if err != nil {
		assert.Fail("error listing captured emails from Mailhog: ", err.Error())
	}

	// Check if the captured email is as expected
	if len(capturedEmails) != 1 {
		assert.Fail("Expected 1 email, but got %d", len(capturedEmails))
	}
	capturedEmail := capturedEmails[0]

	// Validate sender's information
	assert.Contains(capturedEmail.Content.Headers.From[0], "test.send.email.with.bcc@example.com")

	// Validate recipients
	assert.Equal([]string{"recipient1@example.com, recipient2@example.com"}, capturedEmail.Content.Headers.To)

	// Validate Cced recipients
	assert.Equal([]string{"bccrecipient@example.com"}, capturedEmail.Content.Headers.Bcc)

	// Validate email body
	assert.Contains(capturedEmail.Content.Body, "This is a test email sent from Golang with Bcc.")
}

func TestSendEmailWithMissingRecipient(t *testing.T) {
	assert := assert.New(t)
	hr := Email{}

	// Use a dummy SMTP server for testing (e.g., MailHog)
	input := resources.Input(map[string]interface{}{
		schema.AttributeTypeSenderName:   "Flowpipe",
		schema.AttributeTypeFrom:         "sender@example.com",
		schema.AttributeTypeSmtpUsername: "sender@example.com",
		schema.AttributeTypeSmtpPassword: "",
		schema.AttributeTypeHost:         "localhost",
		schema.AttributeTypePort:         int64(1025),
		schema.AttributeTypeBody:         "This is a test email sent from Golang.",
	})

	_, err := hr.Run(context.Background(), input)
	// No errors
	assert.NotNil(err)

	fpErr := err.(perr.ErrorModel)
	assert.Equal("Email input must define to", fpErr.Detail)
	assert.Equal(400, fpErr.Status)
}

func TestSendEmailWithEmptyRecipient(t *testing.T) {
	assert := assert.New(t)
	hr := Email{}

	// Use a dummy SMTP server for testing (e.g., MailHog)
	input := resources.Input(map[string]interface{}{
		schema.AttributeTypeSenderName:   "Flowpipe",
		schema.AttributeTypeFrom:         "sender@example.com",
		schema.AttributeTypeSmtpUsername: "sender@example.com",
		schema.AttributeTypeSmtpPassword: "",
		schema.AttributeTypeHost:         "localhost",
		schema.AttributeTypePort:         int64(1025),
		schema.AttributeTypeTo:           []string{},
		schema.AttributeTypeBody:         "This is a test email sent from Golang.",
	})

	_, err := hr.Run(context.Background(), input)
	// No errors
	assert.NotNil(err)

	fpErr := err.(perr.ErrorModel)
	assert.Equal("Recipients must not be empty", fpErr.Detail)
	assert.Equal(400, fpErr.Status)
}

func TestInvalidPortInput(t *testing.T) {
	assert := assert.New(t)
	hr := Email{}

	// Use a dummy SMTP server for testing (e.g., MailHog)
	input := resources.Input(map[string]interface{}{
		schema.AttributeTypeSenderName:   "TestSendEmail",
		schema.AttributeTypeFrom:         "test.send.email@example.com",
		schema.AttributeTypeSmtpUsername: "test.send.email@example.com",
		schema.AttributeTypeSmtpPassword: "",
		schema.AttributeTypeHost:         "localhost",
		schema.AttributeTypePort:         int64(10000000),
		schema.AttributeTypeTo:           []string{"recipient1@example.com", "recipient2@example.com"},
		schema.AttributeTypeSubject:      "Flowpipe mail test",
		schema.AttributeTypeBody:         "This is a test email sent from Golang.",
	})

	_, err := hr.Run(context.Background(), input)
	assert.NotNil(err)
	assert.Equal(fmt.Sprintf("Bad Request: %d is not a valid port", int64(10000000)), err.Error())
}

func TestEmailInvalidCreds(t *testing.T) {
	assert := assert.New(t)
	hr := Email{}

	input := resources.Input(map[string]interface{}{
		schema.AttributeTypeSenderName:   "Flowpipe",
		schema.AttributeTypeFrom:         "test@example.com",
		schema.AttributeTypeSmtpUsername: "test@example.com",
		schema.AttributeTypeSmtpPassword: "abcdefghijklmnop",
		schema.AttributeTypeHost:         "smtp.gmail.com",
		schema.AttributeTypePort:         int64(587),
		schema.AttributeTypeTo:           []string{"recipient@example.com"},
		schema.AttributeTypeSubject:      "Flowpipe mail test",
		schema.AttributeTypeBody:         "This is a test email message to validate whether the code works or not.",
	})

	output, err := hr.Run(context.Background(), input)
	// No errors
	assert.Nil(err)

	output.HasErrors()
	for _, e := range output.Errors {
		assert.Equal(503, e.Error.Status)
		assert.Equal("535", e.Error.Detail[0:3])
		assert.Contains(e.Error.Detail, "Username and Password not accepted")
	}
}

type HTTPResponse struct {
	Items []CapturedEmail `json:"items"`
}

type CapturedEmail struct {
	Content MailContent `json:"Content"`
}

type MailContent struct {
	Headers Headers `json:"Headers"`
	Body    string  `json:"Body"`
}

type Headers struct {
	Cc   []string `json:"Cc"`
	Bcc  []string `json:"Bcc"`
	To   []string `json:"To"`
	From []string `json:"From"`
	Body string   `json:"body"`
}

func captureEmailsFromSMTP(from string) ([]CapturedEmail, error) {
	// MailHog's API base URL
	mailHogURL := "http://localhost:8025"
	apiEndpoint := "/api/v2/search"
	query := fmt.Sprintf("?kind=containing&query=%s", from)

	// Get a list of all emails received by MailHog
	resp, err := http.Get(mailHogURL + apiEndpoint + query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var v HTTPResponse
	err = json.Unmarshal(body, &v)
	if err != nil {
		return nil, err
	}

	if len(v.Items) == 0 {
		return nil, perr.NotFoundWithMessage("No emails captured in the inbox")
	}

	return v.Items, nil
}
