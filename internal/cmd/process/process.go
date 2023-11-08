package process

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/hokaccha/go-prettyjson"
	"github.com/karrick/gows"
	"github.com/spf13/cobra"
	flowpipeapi "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/flowpipe/internal/cmd/common"
	"github.com/turbot/flowpipe/internal/printers"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/utils"
)

func ProcessCmd(ctx context.Context) (*cobra.Command, error) {

	processCmd := &cobra.Command{
		Use:   "process",
		Short: "Process commands",
	}

	processGetCmd, err := ProcessGetCmd(ctx)
	if err != nil {
		return nil, err
	}
	processCmd.AddCommand(processGetCmd)

	processListCmd, err := ProcessListCmd(ctx)
	if err != nil {
		return nil, err
	}
	processCmd.AddCommand(processListCmd)

	processLogCmd, err := ProcessLogCmd(ctx)
	if err != nil {
		return nil, err
	}
	processCmd.AddCommand(processLogCmd)

	return processCmd, nil

}

func ProcessLogCmd(ctx context.Context) (*cobra.Command, error) {
	var processLogCmd = &cobra.Command{
		Use:  "log <execution-id>",
		Args: cobra.ExactArgs(1),
		Run:  logProcessFunc(ctx),
	}
	return processLogCmd, nil
}

func ProcessGetCmd(ctx context.Context) (*cobra.Command, error) {

	var processGetCmd = &cobra.Command{
		Use:  "get <execution-id>",
		Args: cobra.ExactArgs(1),
		Run:  getProcessFunc(ctx),
	}

	processGetCmd.Flags().BoolP("output-only", "", false, "Get pipeline execution output only")

	return processGetCmd, nil
}

func ProcessListCmd(ctx context.Context) (*cobra.Command, error) {

	var processGetCmd = &cobra.Command{
		Use:  "list",
		Args: cobra.NoArgs,
		Run:  listProcessFunc(ctx),
	}

	return processGetCmd, nil
}

func getStepIcon(name string, failed bool) string {
	if failed {
		return "🔴" // fmt.Sprintf("❌%s", icon)
	}
	icon := " "
	switch name {
	case "http":
		icon = "🔗"
	case "echo":
		icon = "🔠"
	case "pipeline":
		icon = "♊️"
	case "sleep":
		icon = "⏳"
	}
	return icon
}

func getProcessFunc(ctx context.Context) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {

		apiClient := common.GetApiClient()

		outputOnly, _ := cmd.Flags().GetBool("output-only")

		if outputOnly {
			output, _, err := apiClient.ProcessApi.GetOutput(ctx, args[0]).Execute()
			if err != nil {
				error_helpers.ShowError(ctx, err)
				return
			}

			s, err := prettyjson.Marshal(output)

			if err != nil {
				error_helpers.ShowErrorWithMessage(ctx, err, "Error when calling `colorjson.Marshal`")
				return
			}

			fmt.Println(string(s)) //nolint:forbidigo // console output, but we may change it to a different formatter in the future
		} else {
			ex, _, err := apiClient.ProcessApi.Get(ctx, args[0]).Execute()
			if err != nil {
				error_helpers.ShowError(ctx, err)
				return
			}

			s, err := prettyjson.Marshal(ex)

			if err != nil {
				error_helpers.ShowErrorWithMessage(ctx, err, "Error when calling `colorjson.Marshal`")
				return
			}

			fmt.Println(string(s)) //nolint:forbidigo // console output, but we may change it to a different formatter in the future
		}
	}
}

func listProcessFunc(ctx context.Context) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		limit := int32(25) // int32 | The max number of items to fetch per page of data, subject to a min and max of 1 and 100 respectively. If not specified will default to 25. (optional) (default to 25)
		nextToken := ""    // string | When list results are truncated, next_token will be returned, which is a cursor to fetch the next page of data. Pass next_token to the subsequent list request to fetch the next page of data. (optional)

		apiClient := common.GetApiClient()

		processes, _, err := apiClient.ProcessApi.List(ctx).Limit(limit).NextToken(nextToken).Execute()
		if err != nil {
			error_helpers.ShowError(ctx, err)
			return
		}

		if processes != nil {
			printer := printers.GetPrinter(cmd)

			printableResource := types.PrintableProcess{}
			printableResource.Items, err = printableResource.Transform(processes)
			if err != nil {
				error_helpers.ShowErrorWithMessage(ctx, err, "Error when transforming")
			}

			err := printer.PrintResource(ctx, printableResource, cmd.OutOrStdout())
			if err != nil {
				error_helpers.ShowErrorWithMessage(ctx, err, "Error when printing")
			}
		}
	}
}

func logProcessFunc(ctx context.Context) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		apiClient := common.GetApiClient()

		execution, _, err := apiClient.ProcessApi.GetExecution(ctx, args[0]).Execute()
		if err != nil {
			error_helpers.ShowError(ctx, err)
			return
		}

		cols, _, err := gows.GetWinSize()
		if err != nil {
			error_helpers.ShowError(cmd.Context(), err)
			return
		}

		pe := parseExecution(execution)

		lines := []string{fmt.Sprintf("Execution Id: %s", pe.id)}
		for _, plKey := range pe.outerKeys {
			lines = append(lines, renderPipelineExecution(pe.pipelines[plKey], 0, cols)...)
			lines = append(lines, renderPipelineOutput(pe.pipelines[plKey].output, cols)...)
		}

		fmt.Println(strings.Join(lines, "\n")) //nolint:forbidigo // CLI console output
	}
}

func parseExecution(execution *flowpipeapi.ExecutionExecution) parsedExecution {
	var pe parsedExecution
	pe.pipelines = make(map[string]parsedPipeline)
	pe.id = *execution.Id

	for k, v := range *execution.PipelineExecutions {
		if !v.HasParentExecutionId() && !v.HasParentStepExecutionId() {
			pe.outerKeys = append(pe.outerKeys, k)
		}

		pe.pipelines[k] = parsePipeline(&v)
	}

	// Add children to parents
	for key, _ := range pe.pipelines {
		for k, v := range pe.pipelines {
			if v.parentExecutionId != nil && *v.parentExecutionId == key {
				pe.pipelines[key].childPipelines[k] = &v
			}
		}
	}

	return pe
}

func parsePipeline(input *flowpipeapi.ExecutionPipelineExecution) parsedPipeline {
	var pp parsedPipeline

	pp.id = *input.Id
	pp.name = *input.Name
	if input.StartTime != nil {
		pp.startTime, _ = time.Parse(time.RFC3339, *input.StartTime)
	}
	if input.EndTime != nil {
		pp.endTime, _ = time.Parse(time.RFC3339, *input.EndTime)
	}
	pp.output = input.PipelineOutput

	pp.steps = parseStepStatuses(input.StepStatus)
	pp.errorCount = len(input.Errors)

	if input.HasParentExecutionId() {
		pp.parentExecutionId = input.ParentExecutionId
	}

	if input.HasParentStepExecutionId() {
		pp.parentStepExecutionId = input.ParentStepExecutionId
	}

	pp.childPipelines = make(map[string]*parsedPipeline)
	return pp
}

func parseStepStatuses(input *map[string]map[string]flowpipeapi.ExecutionStepStatus) []parsedStepStatus {
	var output []parsedStepStatus

	for stepName, feMap := range *input {
		var pss parsedStepStatus
		pss.name = stepName
		pss.executions = make(map[string][]parsedStepExecution)
		pss.earliestStartTime = time.Now()
		for feKey, execStepStatus := range feMap {
			var pse []parsedStepExecution
			for _, se := range execStepStatus.StepExecutions {
				pse = append(pse, parseStepExecution(&se))
			}
			// Sort Executions by time
			sort.Slice(pse, func(i, j int) bool {
				return pse[i].startTime.Before(pse[j].startTime)
			})

			var earliestStartTime time.Time
			if len(pse) > 0 {
				earliestStartTime = pse[0].startTime
			}
			if earliestStartTime.Before(pss.earliestStartTime) {
				pss.earliestStartTime = earliestStartTime
			}
			pss.executions[feKey] = pse

		}
		output = append(output, pss)
	}

	sort.Slice(output, func(i, j int) bool {
		return output[i].earliestStartTime.Before(output[j].earliestStartTime)
	})

	return output
}

func parseStepExecution(input *flowpipeapi.ExecutionStepExecution) parsedStepExecution {
	var pse parsedStepExecution
	var op flowpipeapi.ExecutionStepExecutionOutput
	pse.id = *input.Id
	pse.name = *input.Name
	pse.status = *input.Status
	if input.HasOutput() {
		op = *input.Output
	}
	pse.output = op.Data
	for _, e := range op.Errors {
		pse.errors = append(pse.errors, *e.Error)
	}
	if input.HasStartTime() {
		pse.startTime, _ = time.Parse(time.RFC3339, *input.StartTime)
	}
	if input.HasEndTime() {
		pse.endTime, _ = time.Parse(time.RFC3339, *input.EndTime)
	}
	return pse
}

type parsedExecution struct {
	id        string
	pipelines map[string]parsedPipeline
	outerKeys []string
}

type parsedPipeline struct {
	id                    string
	name                  string
	output                map[string]any
	startTime             time.Time
	endTime               time.Time
	steps                 []parsedStepStatus
	errorCount            int
	parentExecutionId     *string
	parentStepExecutionId *string
	childPipelines        map[string]*parsedPipeline
}

type parsedStepStatus struct {
	name              string
	earliestStartTime time.Time
	executions        map[string][]parsedStepExecution
}

type parsedStepExecution struct {
	id        string
	name      string
	startTime time.Time
	endTime   time.Time
	output    map[string]any
	errors    []flowpipeapi.PerrErrorModel
	status    string
}

func renderPipelineExecution(pl parsedPipeline, level int, width int) []string {
	var out []string
	indent := strings.Repeat(" ", level*2)
	out = append(out, fmt.Sprintf("%s⏩ %s", indent, pl.name))

	for _, s := range pl.steps {
		stepType := strings.Split(s.name, ".")[0]
		if strings.ToLower(stepType) == "pipeline" {
			icon := getStepIcon(stepType, false)
			content := fmt.Sprintf("%s  %s %s", indent, icon, s.name)
			out = append(out, content)
			for _, execs := range s.executions {
				for _, exec := range execs {
					for _, ppl := range pl.childPipelines {
						if *ppl.parentExecutionId == pl.id && *ppl.parentStepExecutionId == exec.id {
							out = append(out, renderPipelineExecution(*ppl, level+2, width)...)
						}
					}

				}
			}
		} else {
			if len(s.executions) == 1 && len(s.executions["0"]) == 1 {
				out = append(out, renderSingleStepExecution(indent, s.executions["0"][0], width)...)
			} else {
				out = append(out, renderMultipleStepExecutions(indent, s, width)...)
			}
		}
	}

	duration := pl.endTime.Sub(pl.startTime)
	durationPrefix := "Success: "
	if pl.errorCount > 0 {
		durationPrefix = fmt.Sprintf("%d Error(s): ", pl.errorCount)
	}
	out = append(out, renderDurationLine(fmt.Sprintf("%s⏹️  %s", indent, pl.name), durationPrefix, duration, width))
	return out
}

func renderPipelineOutput(output map[string]any, width int) []string {
	var lines []string
	delete(output, "errors")
	if len(output) >= 1 {
		lines = append(lines, "\nOutput:")
	}
	for k, v := range output {
		line := fmt.Sprintf("⇒ %s = %v", k, v)
		if utf8.RuneCountInString(line) >= width {
			line = fmt.Sprintf("%s%s", line[0:width-3], "...")
		}
		lines = append(lines, line)
	}

	return lines
}

func renderDurationLine(content, durationPrefix string, duration time.Duration, width int) string {
	iconsUsing2CodePoints := []string{"⏹️"}
	durationString := utils.HumanizeDuration(duration)
	if utf8.RuneCountInString(durationPrefix) > 0 {
		durationString = fmt.Sprintf("%s%s", durationPrefix, durationString)
	}

	contentWidth := utf8.RuneCountInString(content)
	durationWidth := utf8.RuneCountInString(durationString)

	for _, i := range iconsUsing2CodePoints {
		if strings.Contains(content, i) {
			contentWidth -= 2
		}
	}

	dotCount := width - (contentWidth + durationWidth + 3)
	dots := fmt.Sprintf(" %s ", strings.Repeat(".", dotCount))
	return fmt.Sprintf("%s%s%s", content, dots, durationString)
}

func renderSingleStepExecution(indent string, se parsedStepExecution, width int) []string {
	var out []string
	stepType := strings.Split(se.name, ".")[0]
	icon := getStepIcon(stepType, se.status == "failed")
	duration := se.endTime.Sub(se.startTime)
	content := fmt.Sprintf("%s  %s %s", indent, icon, se.name)
	out = append(out, renderDurationLine(content, "", duration, width))
	// Write Errors
	for _, e := range se.errors {
		errText := fmt.Sprintf("%s     └ Error: %s", indent, e.Detail)
		if len(errText) > width {
			errText = fmt.Sprintf("%s%s", errText[0:width-3], "...")
		}
		out = append(out, errText)
	}
	return out
}

func renderMultipleStepExecutions(indent string, ss parsedStepStatus, width int) []string {
	var out []string
	stepType := strings.Split(ss.name, ".")[0]
	icon := getStepIcon(stepType, false) // Failures will be on execution not parent line
	out = append(out, fmt.Sprintf("%s  %s %s", indent, icon, ss.name))

	for forEachKey, execs := range ss.executions {
		if len(execs) == 1 {
			content := fmt.Sprintf("%s     └- %s", indent, forEachKey)
			duration := execs[0].endTime.Sub(execs[0].startTime)
			out = append(out, renderDurationLine(content, "", duration, width))

			for _, e := range execs[0].errors {
				errText := fmt.Sprintf("%s     └ Error: %s", indent, e.Detail)
				if len(errText) > width {
					errText = fmt.Sprintf("%s%s", errText[0:width-3], "...")
				}
				out = append(out, errText)
			}
		} else {
			for i, ex := range execs {
				content := fmt.Sprintf("%s     └- %s ⇒ [%d]", indent, forEachKey, i)
				duration := ex.endTime.Sub(ex.startTime)
				out = append(out, renderDurationLine(content, "", duration, width))

				for _, e := range ex.errors {
					errText := fmt.Sprintf("%s     └ Error: %s", indent, e.Detail)
					if len(errText) > width {
						errText = fmt.Sprintf("%s%s", errText[0:width-3], "...")
					}
					out = append(out, errText)
				}
			}
		}
	}

	return out
}
