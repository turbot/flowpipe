package util

import (
	"context"
	"fmt"
	"html/template"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"os"
	"strings"
	"time"

	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
)

// TODO: Sync up with email step - once the variable names are in sync, we can use the same schema
func RunSendEmail(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {

	// Read sender credential
	senderEmail := input[schema.AttributeTypeUsername].(string)
	senderCredential := input[schema.AttributeTypePassword].(string)
	host := input[schema.AttributeTypeSmtpServer].(string)
	auth := smtp.PlainAuth("", senderEmail, senderCredential, host)

	// Convert port into integer
	var portInt int64
	if port, ok := input[schema.AttributeTypeSmtpPort].(float64); ok {
		portInt = int64(port)
	}
	if port, ok := input[schema.AttributeTypeSmtpPort].(int64); ok {
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

	header["Content-Type"] = "text/html; charset=\"UTF-8\";"
	header["MIME-version"] = "1.0;"

	// Build the full email message
	var message string
	for key, value := range header {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}

	message += getTemplateMessage(input)

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

func getTemplateMessage(input modconfig.Input) string {

	// Read the email template from a file
	templateFile, err := os.ReadFile("templates/approval-template.html")
	if err != nil {
		fmt.Println("Error reading template:", err)
		// return
	}

	// Parse the email template
	tmpl, err := template.New("email").Parse(string(templateFile))
	if err != nil {
		fmt.Println("Error parsing template:", err)
		// return
	}

	data := struct {
		ExecutionID         string
		PipelineExecutionID string
		StepExecutionID     string
		Options             []interface{}
		ResponseUrl         string
		Prompt              string
	}{
		ExecutionID:         input["executionID"].(string),
		PipelineExecutionID: input["pipelineExecutionID"].(string),
		StepExecutionID:     input["stepExecutionID"].(string),
		Options:             input[schema.AttributeTypeOptions].([]interface{}),
		ResponseUrl:         input[schema.AttributeTypeResponseUrl].(string),
		Prompt:              input[schema.AttributeTypePrompt].(string),
	}

	var body strings.Builder
	// input[schema.AttributeTypeOptions].([]interface{})
	err = tmpl.Execute(&body, data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		// return
	}

	tempMessage := body.String()

	return tempMessage
}
