package primitive

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/pipeparser/schema"
)

func TestSendEmail(t *testing.T) {
	assert := assert.New(t)
	hr := Email{}

	// Use a dummy SMTP server for testing (e.g., MailHog)
	input := types.Input(map[string]interface{}{
		schema.AttributeTypeTo:   []string{"recipient1@example.com"},
		schema.AttributeTypeBody: "This is a test email sent from Golang.",
	})

	// Start the MailHog server as a child process.
	cmd := exec.Command("mailhog")
	if err := cmd.Start(); err != nil {
		panic("Failed to start MailHog server: " + err.Error())
	}
	time.Sleep(10 * time.Second)

	_, err := hr.Run(context.Background(), input)
	// No errors
	assert.Nil(err)

	// Get the captured email data from the SMTP server (e.g., MailHog)
	capturedEmails := captureEmailsFromSMTP()

	// Check if the captured email is as expected
	if len(capturedEmails) != 1 {
		assert.Fail("Expected 1 email, but got %d", len(capturedEmails))
	}
	capturedEmail := capturedEmails[0]

	assert.Equal("sender@example.com", capturedEmail.Raw.From)
	assert.Equal("recipient1@example.com", capturedEmail.Raw.To[0])
	assert.Contains(capturedEmail.Raw.Data, "This is a test email sent from Golang.")

	// Stop the MailHog server.
	if err := cmd.Process.Kill(); err != nil {
		panic("Failed to stop MailHog server: " + err.Error())
	}
}

type HTTPResponse struct {
	Items []CapturedEmail `json:"items"`
}

type CapturedEmail struct {
	Raw EmailRaw `json:"Raw"`
}

type EmailRaw struct {
	From string   `json:"From"`
	To   []string `json:"To"`
	Data string   `json:"Data"`
}

func captureEmailsFromSMTP() []CapturedEmail {
	// MailHog's API base URL
	mailHogURL := "http://localhost:8025"

	// Get a list of all emails received by MailHog
	resp, err := http.Get(mailHogURL + "/api/v2/messages")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var v HTTPResponse
	err = json.Unmarshal(body, &v)
	if err != nil {
		return nil
	}

	return v.Items
}
