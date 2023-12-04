package primitive

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
)

// func TestInputStep(t *testing.T) {
// 	ctx := context.Background()
//
// 	assert := assert.New(t)
// 	hr := Input{
// 		ExecutionID:         "exec_cknkhj5gdurd7349d4v0",
// 		StepExecutionID:     "sexec_cknkhj5gdurd7349d510",
// 		PipelineExecutionID: "pexec_cknkhj5gdurd7349d4vg",
// 	}

// 	input := modconfig.Input(map[string]interface{}{
// 		"type": InputTypeSlack,
// 	})

// 	_, err := hr.Run(ctx, input)
// 	assert.Nil(err)
// 	// assert.Equal("200 OK", output.Get("status"))
// 	// assert.Equal(200, output.Get("status_code"))
// 	// assert.Equal("text/html; charset=utf-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
// 	// assert.Contains(output.Get(schema.AttributeTypeResponseBody), "Steampipe")
// }

// func TestIntegrationInputEmailMain(m *testing.M) {
// 	// Start MailHog before running tests
// 	startMailHog()
// 	time.Sleep(2 * time.Second) // Wait for the server to be ready

// 	// Run tests
// 	code := m.Run()

// 	// Stop MailHog after tests are completed
// 	stopMailHog()

// 	// Exit with the test code
// 	os.Exit(code)
// }

func XXXTestIntegrationInputSendEmail(t *testing.T) {
	assert := assert.New(t)
	hr := Input{}

	// Use a dummy SMTP server for testing (e.g., MailHog)
	input := modconfig.Input(map[string]interface{}{
		// schema.AttributeTypeSenderName: "Karan",

		schema.AttributeTypeType:       InputTypeEmail,
		schema.AttributeTypeUsername:   "karan@turbot.com",
		schema.AttributeTypePassword:   "xxxxxx",
		schema.AttributeTypeSmtpServer: "smtp.gmail.com",

		// schema.AttributeTypePort:    int64(587),
		// schema.AttributeTypeTo:      []string{"karan@turbot.com"},
		// schema.AttributeTypeSubject: "Flowpipe mail test",
	})

	_, err := hr.Run(context.Background(), input)
	// No errors
	assert.Nil(err)

	// // Get the captured email data from the SMTP server (e.g., MailHog)
	// capturedEmails, err := captureEmailsFromSMTP(input[schema.AttributeTypeFrom].(string))
	// if err != nil {
	// 	assert.Fail("error listing captured emails from Mailhog: ", err.Error())
	// }

	// // Check if the captured email is as expected
	// if len(capturedEmails) != 1 {
	// 	assert.Fail("Expected 1 email, but got %d", len(capturedEmails))
	// }
	// capturedEmail := capturedEmails[0]

	// // Validate sender's information
	// assert.Contains(capturedEmail.Content.Headers.From[0], "test.send.email@example.com")

	// // Validate recipients
	// assert.Equal([]string{"recipient1@example.com, recipient2@example.com"}, capturedEmail.Content.Headers.To)

	// // Validate email body
	// assert.Contains(capturedEmail.Content.Body, "This is a test email sent from Golang.")
}
