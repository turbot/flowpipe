package primitive

import (
	"bytes"
	"context"
	"fmt"
	"github.com/turbot/pipe-fittings/modconfig"
	"os"
)

type InputIntegrationConsole struct {
	InputIntegrationBase
}

func NewInputIntegrationConsole(base InputIntegrationBase) InputIntegrationConsole {
	return InputIntegrationConsole{
		InputIntegrationBase: base,
	}
}

func (ip *InputIntegrationConsole) PostMessage(_ context.Context, mc MessageCreator, options []InputIntegrationResponseOption) (*modconfig.Output, error) {
	output := modconfig.Output{}

	text, form, response, err := mc.ConsoleMessage(ip, options)
	if err != nil {
		return nil, err
	}

	switch mc.(type) {
	case *MessageStepMessageCreator:
		fmt.Println(*text)
		output.Data = map[string]interface{}{"value": text}
		output.Status = "finished"
	case *InputStepMessageCreator:
		is := mc.(*InputStepMessageCreator)
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() {
			_ = w.Close()
			os.Stdout = oldStdout

			// Print the buffered output
			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			fmt.Print(buf.String())
		}()

		if err := form.Run(); err != nil {
			return nil, err
		}

		output.Data = map[string]interface{}{"value": response}
		output.Status = "finished"
		fmt.Printf("Prompt: %s\n", is.Prompt)
		fmt.Printf("Response: '%s'\n", *response.(*string))
	}

	return &output, nil
}
