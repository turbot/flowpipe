package primitive

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/pipeparser/schema"
)

func TestSendEmail(t *testing.T) {
	assert := assert.New(t)
	hr := Email{
		Setting: "mailhog",
	}

	// Use a dummy SMTP server for testing (e.g., MailHog)
	input := types.Input(map[string]interface{}{
		schema.AttributeTypeTo:   []string{"recipient1@example.com", "recipient2@example.com"},
		schema.AttributeTypeBody: "This is a test email sent from Golang.",
	})

	// Start the MailHog server as a child process.
	cmd := exec.Command("mailhog")
	if err := cmd.Start(); err != nil {
		panic("Failed to start MailHog server: " + err.Error())
	}
	time.Sleep(2 * time.Second) // Wait for the server to be ready

	_, err := hr.Run(context.Background(), input)
	// No errors
	assert.Nil(err)

	// Get the captured email data from the SMTP server (e.g., MailHog)
	capturedEmails, err := captureEmailsFromSMTP()
	if err != nil {
		assert.Fail("error listing captured emails from Mailhog: ", err.Error())
	}

	// Check if the captured email is as expected
	if len(capturedEmails) != 1 {
		assert.Fail("Expected 1 email, but got %d", len(capturedEmails))
	}
	capturedEmail := capturedEmails[0]

	// Validate sender's information
	assert.Contains(capturedEmail.Content.Headers.From[0], "sender@example.com")

	// Validate recipients
	assert.Equal([]string{"recipient1@example.com, recipient2@example.com"}, capturedEmail.Content.Headers.To)

	// Validate email body
	assert.Contains(capturedEmail.Content.Body, "This is a test email sent from Golang.")

	// Stop the MailHog server.
	if err := cmd.Process.Kill(); err != nil {
		panic("Failed to stop MailHog server: " + err.Error())
	}
}

func TestSendEmailWithCc(t *testing.T) {
	assert := assert.New(t)
	hr := Email{
		Setting: "mailhog",
	}

	// Use a dummy SMTP server for testing (e.g., MailHog)
	input := types.Input(map[string]interface{}{
		schema.AttributeTypeTo:   []string{"recipient1@example.com", "recipient2@example.com"},
		schema.AttributeTypeCc:   []string{"ccrecipient@example.com"},
		schema.AttributeTypeBody: "This is a test email sent from Golang with Cc.",
	})

	// Start the MailHog server as a child process.
	cmd := exec.Command("mailhog")
	if err := cmd.Start(); err != nil {
		panic("Failed to start MailHog server: " + err.Error())
	}
	time.Sleep(2 * time.Second) // Wait for the server to be ready

	_, err := hr.Run(context.Background(), input)
	// No errors
	assert.Nil(err)

	// Get the captured email data from the SMTP server (e.g., MailHog)
	capturedEmails, err := captureEmailsFromSMTP()
	if err != nil {
		assert.Fail("error listing captured emails from Mailhog: ", err.Error())
	}

	// Check if the captured email is as expected
	if len(capturedEmails) != 1 {
		assert.Fail("Expected 1 email, but got %d", len(capturedEmails))
	}
	capturedEmail := capturedEmails[0]

	// Validate sender's information
	assert.Contains(capturedEmail.Content.Headers.From[0], "sender@example.com")

	// Validate recipients
	assert.Equal([]string{"recipient1@example.com, recipient2@example.com"}, capturedEmail.Content.Headers.To)

	// Validate Cced recipients
	assert.Equal([]string{"ccrecipient@example.com"}, capturedEmail.Content.Headers.Cc)

	// Validate email body
	assert.Contains(capturedEmail.Content.Body, "This is a test email sent from Golang with Cc.")

	// Stop the MailHog server.
	if err := cmd.Process.Kill(); err != nil {
		panic("Failed to stop MailHog server: " + err.Error())
	}
}

func TestSendEmailWithBcc(t *testing.T) {
	assert := assert.New(t)
	hr := Email{
		Setting: "mailhog",
	}

	// Use a dummy SMTP server for testing (e.g., MailHog)
	input := types.Input(map[string]interface{}{
		schema.AttributeTypeTo:   []string{"recipient1@example.com", "recipient2@example.com"},
		schema.AttributeTypeBcc:  []string{"bccrecipient@example.com"},
		schema.AttributeTypeBody: "This is a test email sent from Golang with Bcc.",
	})

	// Start the MailHog server as a child process.
	cmd := exec.Command("mailhog")
	if err := cmd.Start(); err != nil {
		panic("Failed to start MailHog server: " + err.Error())
	}
	time.Sleep(2 * time.Second) // Wait for the server to be ready

	_, err := hr.Run(context.Background(), input)
	// No errors
	assert.Nil(err)

	// Get the captured email data from the SMTP server (e.g., MailHog)
	capturedEmails, err := captureEmailsFromSMTP()
	if err != nil {
		assert.Fail("error listing captured emails from Mailhog: ", err.Error())
	}

	// Check if the captured email is as expected
	if len(capturedEmails) != 1 {
		assert.Fail("Expected 1 email, but got %d", len(capturedEmails))
	}
	capturedEmail := capturedEmails[0]

	// Validate sender's information
	assert.Contains(capturedEmail.Content.Headers.From[0], "sender@example.com")

	// Validate recipients
	assert.Equal([]string{"recipient1@example.com, recipient2@example.com"}, capturedEmail.Content.Headers.To)

	// Validate Cced recipients
	assert.Equal([]string{"bccrecipient@example.com"}, capturedEmail.Content.Headers.Bcc)

	// Validate email body
	assert.Contains(capturedEmail.Content.Body, "This is a test email sent from Golang with Bcc.")

	// Stop the MailHog server.
	if err := cmd.Process.Kill(); err != nil {
		panic("Failed to stop MailHog server: " + err.Error())
	}
}

func TestSendEmailWithMissingRecipient(t *testing.T) {
	assert := assert.New(t)
	hr := Email{}

	// Use a dummy SMTP server for testing (e.g., MailHog)
	input := types.Input(map[string]interface{}{
		schema.AttributeTypeBody: "This is a test email sent from Golang.",
	})

	_, err := hr.Run(context.Background(), input)
	// No errors
	assert.NotNil(err)

	fpErr := err.(fperr.ErrorModel)
	assert.Equal("Email input must define a recipients", fpErr.Detail)
	assert.Equal(400, fpErr.Status)
}

func TestSendEmailWithEmptyRecipient(t *testing.T) {
	assert := assert.New(t)
	hr := Email{}

	// Use a dummy SMTP server for testing (e.g., MailHog)
	input := types.Input(map[string]interface{}{
		schema.AttributeTypeTo:   []string{},
		schema.AttributeTypeBody: "This is a test email sent from Golang.",
	})

	_, err := hr.Run(context.Background(), input)
	// No errors
	assert.NotNil(err)

	fpErr := err.(fperr.ErrorModel)
	assert.Equal("Recipients must not be empty", fpErr.Detail)
	assert.Equal(400, fpErr.Status)
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

func captureEmailsFromSMTP() ([]CapturedEmail, error) {
	// MailHog's API base URL
	mailHogURL := "http://localhost:8025"

	// Get a list of all emails received by MailHog
	resp, err := http.Get(mailHogURL + "/api/v2/messages")
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
		return nil, fmt.Errorf("No emails captured in the inbox")
	}

	return v.Items, nil
}
