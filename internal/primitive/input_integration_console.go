//nolint:forbidigo // console output, will need fmt.Println()
package primitive

import (
	"bytes"
	"context"
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
	o "github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/modconfig"
	"os"
	"strings"
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

	o.PipelineProgress.Stop()

	switch m := mc.(type) {
	case *MessageStepMessageCreator:
		fmt.Println(*text)
		output.Data = map[string]interface{}{"value": text}
		output.Status = "finished"
	case *InputStepMessageCreator:
		var theme *huh.Theme
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
		var displayResponse string
		switch v := response.(type) {
		case *[]string:
			if !helpers.IsNil(v) {
				displayResponse = strings.Join(*v, ", ")
			}
		case *string:
			displayResponse = *v
		}
		if enableColor {
			fmt.Printf("%s\n", m.Prompt)
			fmt.Printf("%s\n\n", lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#006400", Dark: "#00FF00"}).Render(displayResponse))
		} else {
			fmt.Printf("%s\n", m.Prompt)
			fmt.Printf("%s\n\n", displayResponse)
		}
	}

	o.PipelineProgress.Start()
	return &output, nil
}
