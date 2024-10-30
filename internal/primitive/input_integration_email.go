package primitive

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/smtp"
	"net/textproto"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/flowpipe/internal/templates"
	kitTypes "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

type InputIntegrationEmailMessage interface {
	Message() (string, error)
}

type InputIntegrationEmail struct {
	InputIntegrationBase

	Host       *string
	Port       *int64
	SecurePort *int64
	Tls        *string
	To         []string
	Cc         []string
	Bcc        []string
	From       string
	Subject    string
	User       *string
	Pass       *string
	FormUrl    string
}

func NewInputIntegrationEmail(base InputIntegrationBase) InputIntegrationEmail {
	return InputIntegrationEmail{
		InputIntegrationBase: base,
	}
}

func (ip *InputIntegrationEmail) ValidateInputIntegrationEmail(ctx context.Context, i resources.Input) error {

	// Validate sender's information
	if i[schema.AttributeTypeFrom] == nil {
		return perr.BadRequestWithMessage("Email input must define from")
	}
	if i[schema.AttributeTypeSmtpPassword] == nil {
		return perr.BadRequestWithMessage("Email input must define sender_credential")
	}
	if i[schema.AttributeTypeHost] == nil {
		return perr.BadRequestWithMessage("Email input must define a SMTP host")
	}
	if i[schema.AttributeTypePort] == nil {
		return perr.BadRequestWithMessage("Email input must define a port")
	}

	// Validate the port input
	if i[schema.AttributeTypePort] != nil {
		var port int64
		switch data := i[schema.AttributeTypePort].(type) {
		case float64:
			port = int64(data)
		case int64:
			port = data
		default:
			return perr.BadRequestWithMessage("port must be a number")
		}

		portInString := strconv.FormatInt(port, 10)
		match, err := regexp.MatchString("^((6553[0-5])|(655[0-2][0-9])|(65[0-4][0-9]{2})|(6[0-4][0-9]{3})|([1-5][0-9]{4})|([0-5]{0,5})|([0-9]{1,4}))$", portInString)
		if err != nil {
			return perr.BadRequestWithMessage("error while validating the port")
		}

		if !match {
			return perr.BadRequestWithMessage(fmt.Sprintf("%s is not a valid port", portInString))
		}
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

func (ip *InputIntegrationEmail) PostMessage(ctx context.Context, mc MessageCreator, options []InputIntegrationResponseOption) (*resources.Output, error) {
	var err error
	host := kitTypes.SafeString(ip.Host)
	tls := kitTypes.SafeString(ip.Tls)
	var port int64
	switch tls {
	case "off":
		if ip.Port != nil {
			port = *ip.Port
		} else {
			port = 25 // default
		}
	case "required":
		if ip.SecurePort != nil {
			port = *ip.SecurePort
		} else {
			port = 587 // default
		}
	default: // default is auto, this should also be used if unset, defaults to port 587
		if ip.SecurePort != nil {
			port = *ip.SecurePort
		} else if ip.Port != nil {
			port = *ip.Port
		} else {
			port = 587
		}
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	auth := smtp.PlainAuth("", kitTypes.SafeString(ip.User), kitTypes.SafeString(ip.Pass), host)

	output := resources.Output{
		Data: map[string]interface{}{},
	}

	message, err := mc.EmailMessage(ip, options)
	if err != nil {
		return nil, perr.InternalWithMessage(fmt.Sprintf("unable to create email message: %s", err.Error()))
	}

	recipients := ip.To
	if len(ip.Cc) > 0 {
		recipients = append(recipients, ip.Cc...)
	}
	if len(ip.Bcc) > 0 {
		recipients = append(recipients, ip.Bcc...)
	}

	start := time.Now().UTC()
	err = smtp.SendMail(addr, auth, ip.From, recipients, []byte(message))
	finish := time.Now().UTC()

	output.Flowpipe = FlowpipeMetadataOutput(start, finish)

	if err != nil {
		var smtpError *textproto.Error
		if !errors.As(err, &smtpError) {
			return nil, perr.InternalWithMessage(fmt.Sprintf("unable to send email: %s", err.Error()))
		}
		switch {
		case smtpError.Code >= 400 && smtpError.Code <= 499:
			return &output, perr.BadRequestWithMessage(fmt.Sprintf("unable to send email: %d %s", smtpError.Code, smtpError.Msg))
		case smtpError.Code >= 500 && smtpError.Code <= 599:
			return &output, perr.ServiceUnavailableWithMessage(fmt.Sprintf("unable to send email: %d %s", smtpError.Code, smtpError.Msg))
		default:
			return &output, perr.InternalWithMessage(fmt.Sprintf("unable to send email: %d %s", smtpError.Code, smtpError.Msg))
		}
	}

	return &output, nil
}

func parseEmailInputTemplate(templateFileName string, data any) (string, error) {
	funcs := template.FuncMap{
		"mod": func(a, b int) int { return a % b },
		"sub": func(a, b int) int { return a - b },
		"getColor": func(s *string) string {
			if s != nil {
				switch *s {
				case "ok":
					return "#379634"
				case "alert":
					return "#b32128"
				default:
					return "#036"
				}
			}
			return "#036"
		},
	}
	templateFile, err := templates.HTMLTemplate(templateFileName)
	if err != nil {
		return "", perr.InternalWithMessage("error while reading the email template")
	}
	tmpl, err := template.New("email").Funcs(funcs).Parse(string(templateFile))
	if err != nil {
		slog.Error("error while parsing the email template", "error", err)
		return "", err
	}

	var body strings.Builder
	err = tmpl.Execute(&body, data)
	if err != nil {
		slog.Error("error while executing the email template", "error", err)
		return "", perr.BadRequestWithMessage("error while executing the email template")
	}

	tempMessage := body.String()

	return tempMessage, nil
}
