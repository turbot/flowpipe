package primitive

import (
	"bytes"
	"context"
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
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
		var theme *huh.Theme
		is := mc.(*InputStepMessageCreator)
		enableColor := viper.GetString(constants.ArgOutput) == constants.OutputFormatPretty

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

		if enableColor {
			theme = huh.ThemeCharm()
		} else {
			theme = huh.ThemeBase()
		}
		if err := form.WithTheme(theme).Run(); err != nil {
			return nil, err
		}

		output.Data = map[string]interface{}{"value": response}
		output.Status = "finished"
		if enableColor {
			fmt.Printf("%s: %s\n", is.Prompt, lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#006400", Dark: "#00FF00"}).Render(*response.(*string)))
		} else {
			fmt.Printf("%s: %s\n", is.Prompt, *response.(*string))
		}
	}

	return &output, nil
}
