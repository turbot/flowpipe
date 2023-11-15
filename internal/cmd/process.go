package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
	"reflect"
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

// process commands
func processCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "process",
		Short: "Process commands",
	}

	cmd.AddCommand(processGetCmd())
	cmd.AddCommand(processListCmd())
	cmd.AddCommand(processLogCmd())

	return cmd

}

// get
func processGetCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:  "get <execution-id>",
		Args: cobra.ExactArgs(1),
		Run:  getProcessFunc,
	}

	cmdconfig.
		OnCmd(cmd).
		AddBoolFlag(constants.ArgOutputOnly, false, "Get pipeline execution output only")

	return cmd
}

func getProcessFunc(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
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

// log
func processLogCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:  "log <execution-id>",
		Args: cobra.ExactArgs(1),
		Run:  logProcessFunc,
	}
	return cmd
}

func logProcessFunc(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
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
		lines = append(lines, renderPipeline(pe.pipelines[plKey], 0, cols)...)
		lines = append(lines, renderPipelineOutput(pe.pipelines[plKey].output, cols)...)
	}

	fmt.Println(strings.Join(lines, "\n")) //nolint:forbidigo // CLI console output
}

// list
func processListCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:  "list",
		Args: cobra.NoArgs,
		Run:  listProcessFunc,
	}

	return cmd
}

func listProcessFunc(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
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

func getIndentByLevel(level int) string {
	return strings.Repeat(" ", level*2)
}

func getIcon(name string) string {
	icon := "❓"
	switch name {
	case "http":
		icon = "🔗"
	case "echo":
		icon = "🔠"
	case "pipeline":
		icon = "♊️"
	case "sleep":
		icon = "⏳"
	case "failed":
		icon = "🔴"
	case "finished":
		icon = "✅"
	}
	return icon
}

func parseExecution(execution *flowpipeapi.ExecutionExecution) parsedExecution {
	var pe parsedExecution
	pe.pipelines = make(map[string]parsedPipeline)
	pe.id = *execution.Id

	plex := *execution.PipelineExecutions
	for k := range plex {
		v := plex[k]
		if !v.HasParentExecutionId() && !v.HasParentStepExecutionId() {
			pe.outerKeys = append(pe.outerKeys, k)
		}
		pe.pipelines[k] = parsePipeline(&v)
	}

	// Add children to parents
	for k, v := range *execution.PipelineExecutions {
		if v.HasParentExecutionId() && v.HasParentStepExecutionId() {
			parentId := *v.ParentExecutionId
			parentStepExecId := *v.ParentStepExecutionId
			child := pe.pipelines[k]
			pe.pipelines[parentId].childPipelines[parentStepExecId] = &child
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
			for i := range execStepStatus.StepExecutions {
				se := execStepStatus.StepExecutions[i]
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

func renderPipeline(pl parsedPipeline, level int, width int) []string {
	var out []string
	indent := getIndentByLevel(level)
	out = append(out, fmt.Sprintf("%s⏩ %s", indent, pl.name))
	out = append(out, renderPipelineSteps(pl.steps, &pl, level+1, width)...)

	duration := pl.endTime.Sub(pl.startTime)
	durationPrefix := "Success: "
	if pl.errorCount > 0 {
		durationPrefix = fmt.Sprintf("%d Error(s): ", pl.errorCount)
	}
	out = append(out, renderDurationLine(fmt.Sprintf("%s⏹️  %s", indent, pl.name), durationPrefix, duration, width))
	return out
}

func renderPipelineSteps(steps []parsedStepStatus, parent *parsedPipeline, level int, width int) []string {
	var out []string
	indent := getIndentByLevel(level)

	for _, step := range steps {
		stepType := strings.Split(step.name, ".")[0]
		stepIcon := getIcon(stepType)
		out = append(out, fmt.Sprintf("%s%s %s", indent, stepIcon, step.name))
		out = append(out, renderPipelineStepExecutions(stepType, step.executions, parent, level+1, width)...)
	}

	return out
}

func renderPipelineStepExecutions(stepType string, execs map[string][]parsedStepExecution, parent *parsedPipeline, level, width int) []string {
	var out []string

	if strings.ToLower(stepType) == "pipeline" {
		for _, exec := range execs {
			for _, e := range exec {
				if parent.childPipelines[e.id] != nil {
					cpl := *parent.childPipelines[e.id]
					out = append(out, renderPipeline(cpl, level+1, width)...)
				}
			}
		}
	} else {
		indent := getIndentByLevel(level)
		if len(execs) == 1 && len(execs["0"]) == 1 {
			out = append(out, renderSingleStepExecution(indent, execs["0"][0], width)...)
		} else {
			for forEachKey, forEachValues := range execs {
				if len(execs) == 1 {
					forEachKey = ""
				}
				out = append(out, renderMultipleStepExecutions(getIndentByLevel(level), forEachKey, forEachValues, width)...)
			}
		}
	}

	return out
}

func renderPipelineOutput(output map[string]any, width int) []string {
	var lines []string
	delete(output, "errors")
	if len(output) >= 1 {
		lines = append(lines, "\nOutput:")
	}
	for k, v := range output {
		var line string
		if !isSimpleType(v) {
			jsonData, err := json.Marshal(v)
			if err != nil {
				line = fmt.Sprintf("⇒ %s - Error rendering output value.", k)
			} else {
				line = fmt.Sprintf("⇒ %s = %s", k, string(jsonData))
			}
		} else {
			line = fmt.Sprintf("⇒ %s = %v", k, v)
		}

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
	dots := strings.Repeat(".", dotCount)
	return fmt.Sprintf("%s %s %s", content, dots, durationString)
}

func renderSingleStepExecution(indent string, se parsedStepExecution, width int) []string {
	var out []string
	icon := getIcon(se.status)
	duration := se.endTime.Sub(se.startTime)
	content := fmt.Sprintf("%s%s %s", indent, icon, se.name)
	out = append(out, renderDurationLine(content, "", duration, width))
	// Write Errors
	for _, e := range se.errors {
		errText := fmt.Sprintf("%s   └ Error: %s", indent, e.Detail)
		if len(errText) > width {
			errText = fmt.Sprintf("%s%s", errText[0:width-3], "...")
		}
		out = append(out, errText)
	}
	return out
}

func renderMultipleStepExecutions(indent string, key string, ses []parsedStepExecution, width int) []string {
	var out []string

	for i, se := range ses {
		icon := getIcon(se.status)
		counter := ""
		if len(ses) > 1 {
			counter = fmt.Sprintf("[%d]", i+1)
		}
		content := fmt.Sprintf("%s%s %s%s", indent, icon, key, counter)
		duration := se.endTime.Sub(se.startTime)
		out = append(out, renderDurationLine(content, "", duration, width))

		for _, e := range se.errors {
			errText := fmt.Sprintf("%s  └ Error: %s", indent, e.Detail)
			if len(errText) > width {
				errText = fmt.Sprintf("%s%s", errText[0:width-3], "...")
			}
			out = append(out, errText)
		}
	}
	return out
}

func isSimpleType(input any) bool {
	kind := reflect.TypeOf(input).Kind()
	switch kind {
	case
		reflect.Bool,
		reflect.String,
		reflect.Float32,
		reflect.Float64,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		return true
	default:
		return false
	}
}
