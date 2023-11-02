package process

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/hokaccha/go-prettyjson"
	"github.com/karrick/gows"
	"github.com/spf13/cobra"
	flowpipeapi "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/flowpipe/internal/cmd/common"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/printers"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
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

type pipelineExecution struct {
	executionId  string
	pipelineName string
	startTime    *time.Time
	endTime      *time.Time
	cancelled    bool
	steps        []*pipelineStep
	output       map[string]any
}

type pipelineStep struct {
	stepName   string
	startTime  *time.Time
	endTime    *time.Time
	executions []*stepExecution
}

func (ps *pipelineStep) failed() bool {
	for _, sel := range ps.executions {
		if sel.status == "failed" {
			return true
		}
	}
	return false
}

func (ps *pipelineStep) setStartTime(t time.Time) {
	if ps.startTime == nil || ps.startTime.After(t) {
		ps.startTime = &t
	}
}

func (ps *pipelineStep) setEndTime(t time.Time) {
	if ps.endTime == nil || ps.endTime.Before(t) {
		ps.endTime = &t
	}
}

type stepExecution struct {
	execKey            string // the key for for_each -> blank if not for_each step
	stepExecutionId    string
	stepName           string
	stepErrors         []string
	status             string
	output             *modconfig.Output
	startTime          time.Time
	endTime            time.Time
	childPipeline      *pipelineExecution
	parentPipelineStep *pipelineStep
}

func fetchLogsForProcess(ctx context.Context, execId string) ([]flowpipeapi.ProcessEventLog, error) {
	logResponse, _, err := common.GetApiClient().ProcessApi.GetLog(ctx, execId).Execute()
	if err != nil {
		return nil, err
	}
	return logResponse.Items, nil
}

func logProcessFunc(ctx context.Context) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {

		logs, err := fetchLogsForProcess(cmd.Context(), args[0])
		error_helpers.FailOnError(err)

		// a set of lookup maps to make updates easier
		// map keyed {pipeline_execution_id}
		pipelinesExecuted := map[string]*pipelineExecution{}
		// map keyed by {step_execution_id}
		stepsExecuted := map[string]*stepExecution{}
		// map keyed by `{pipeline_execution_id}_{step_name}`
		pipelineSteps := map[string]*pipelineStep{}

		var executionLog *pipelineExecution = &pipelineExecution{}

		for _, logEntry := range logs {
			payload := logEntry.GetPayload()
			eventType := logEntry.GetEventType()

			switch eventType {
			case "handler.pipeline_queued":
				var et event.PipelineQueued
				err := json.Unmarshal([]byte(payload), &et)
				if err != nil {
					error_helpers.ShowError(cmd.Context(), err)
				}
				exec := &pipelineExecution{
					pipelineName: et.Name,
					executionId:  et.PipelineExecutionID,
					steps:        []*pipelineStep{},
				}
				// put it in the map so that we can refer to it later
				// we will be updating the value in the map
				// since the value is a pointer, the desired one will get updated as well
				pipelinesExecuted[et.PipelineExecutionID] = exec

				// now put it in the proper place as well
				if len(et.ParentStepExecutionID) != 0 {
					// this is a child pipeline of another step
					stepsExecuted[et.ParentStepExecutionID].childPipeline = exec
				} else {
					// this is the root pipeline
					executionLog = exec
				}

			case "handler.pipeline_started":
				var et event.PipelineStarted
				err := json.Unmarshal([]byte(payload), &et)
				if err != nil {
					error_helpers.ShowError(cmd.Context(), err)
				}

				pipelinesExecuted[et.PipelineExecutionID].startTime = &et.Event.CreatedAt
			case "handler.pipeline_step_queued":
				var et event.PipelineStepQueued
				err := json.Unmarshal([]byte(payload), &et)
				if err != nil {
					error_helpers.ShowError(cmd.Context(), err)
					return
				}

				keyInPipelineStepsMap := fmt.Sprintf("%s_%s", et.PipelineExecutionID, et.StepName)
				// do we have this in the pipelineSteps map?
				theStep, gotit := pipelineSteps[keyInPipelineStepsMap]
				if !gotit {
					ps := &pipelineStep{
						stepName:   et.StepName,
						executions: []*stepExecution{},
						// assume that this is the start time - it may get overridden by the case "handler.pipeline_step_started"
						startTime: &et.Event.CreatedAt,
					}
					// we have never encountered this step before
					pipelineSteps[keyInPipelineStepsMap] = ps
					pipelinesExecuted[et.PipelineExecutionID].steps = append(pipelinesExecuted[et.PipelineExecutionID].steps, ps)
					theStep = ps
				}

				stepExecLog := &stepExecution{
					stepExecutionId: et.StepExecutionID,
					stepName:        et.StepName,
					// assume that this step was started now - the handler log will overwrite
					startTime:          et.Event.CreatedAt,
					parentPipelineStep: theStep,
				}

				if et.StepForEach != nil {
					stepExecLog.execKey = et.StepForEach.Key
				}

				// this is a pipeline step execution - add it to the pipelineStep
				theStep.executions = append(theStep.executions, stepExecLog)

				// put this in the steps executed map, so that we can refer to it easier
				stepsExecuted[et.StepExecutionID] = stepExecLog
			case "handler.pipeline_step_started":
				var et event.PipelineStepStarted
				err := json.Unmarshal([]byte(payload), &et)
				if err != nil {
					error_helpers.ShowError(cmd.Context(), err)
					return
				}
				stepsExecuted[et.StepExecutionID].startTime = et.Event.CreatedAt
				stepsExecuted[et.StepExecutionID].parentPipelineStep.setStartTime(et.Event.CreatedAt)
			case "handler.pipeline_step_finished":
				var et event.PipelineStepFinished
				err := json.Unmarshal([]byte(payload), &et)
				if err != nil {
					error_helpers.ShowError(cmd.Context(), err)
				}
				stepsExecuted[et.StepExecutionID].endTime = et.Event.CreatedAt
				stepsExecuted[et.StepExecutionID].status = et.Output.Status
				stepsExecuted[et.StepExecutionID].output = et.Output
				stepsExecuted[et.StepExecutionID].stepErrors = []string{}
				for _, se := range et.Output.Errors {
					// add in the errors
					stepsExecuted[et.StepExecutionID].stepErrors = append(stepsExecuted[et.StepExecutionID].stepErrors, se.Error.Detail)
				}

				stepsExecuted[et.StepExecutionID].parentPipelineStep.setEndTime(et.Event.CreatedAt)
			case "handler.pipeline_canceled":
				var et event.PipelineCanceled
				err := json.Unmarshal([]byte(payload), &et)
				if err != nil {
					error_helpers.ShowError(cmd.Context(), err)
				}
				pipelinesExecuted[et.PipelineExecutionID].cancelled = true
				pipelinesExecuted[et.PipelineExecutionID].endTime = &et.Event.CreatedAt
			case "command.pipeline_fail":
				var et event.PipelineFail
				err := json.Unmarshal([]byte(payload), &et)
				if err != nil {
					error_helpers.ShowError(cmd.Context(), err)
				}
				// record the end tim here - sometimes the handlers do not fire apparently - can't rely on it
				pipelinesExecuted[et.PipelineExecutionID].endTime = &et.Event.CreatedAt
			case "handler.pipeline_finished":
				var et event.PipelineFinished
				err := json.Unmarshal([]byte(payload), &et)
				if err != nil {
					error_helpers.ShowError(cmd.Context(), err)
				}
				pipelinesExecuted[et.PipelineExecutionID].endTime = &et.Event.CreatedAt
				pipelinesExecuted[et.PipelineExecutionID].output = et.PipelineOutput
			case "handler.pipeline_failed":
				var et event.PipelineFailed
				err := json.Unmarshal([]byte(payload), &et)
				if err != nil {
					error_helpers.ShowError(cmd.Context(), err)
				}
				pipelinesExecuted[et.PipelineExecutionID].endTime = &et.Event.CreatedAt
				pipelinesExecuted[et.PipelineExecutionID].output = et.PipelineOutput
			default:
				// Ignore unknown types while loading
			}
		}

		cols, _, err := gows.GetWinSize()
		if err != nil {
			error_helpers.ShowError(cmd.Context(), err)
			return
		}

		lines := renderExecutionLog(cmd.Context(), executionLog, 0, cols)

		fmt.Println()                          //nolint:forbidigo // CLI console output
		fmt.Println(strings.Join(lines, "\n")) //nolint:forbidigo // CLI console output
		fmt.Println()                          //nolint:forbidigo // CLI console output
	}
}

func getIndentForLevel(level int) string {
	indent := ""
	for i := 0; i < (level * 2); i++ {
		indent += " "
	}
	return indent
}

func renderExecutionLog(ctx context.Context, log *pipelineExecution, level int, width int) []string {
	lines := []string{}
	indent := getIndentForLevel(level)
	lines = append(lines, fmt.Sprintf("%s⏩ %s", indent, log.pipelineName))
	for _, step := range log.steps {
		lines = append(lines, renderPipelineStep(ctx, step, level, width)...)
	}
	lines = append(lines, renderLineWithDuration(ctx, fmt.Sprintf("%s⏹️  %s", indent, log.pipelineName), log.endTime.Sub(*log.startTime), "Total: ", width))
	// lines = append(lines, fmt.Sprintf("%s⏹️  %s", indent, log.pipelineName))
	lines = append(lines, renderPipelineOutput(ctx, log.output, width)...)
	return lines
}

func renderPipelineStep(ctx context.Context, step *pipelineStep, level int, width int) []string {
	icon := getStepIcon(strings.Split(step.stepName, ".")[0], step.failed())
	indent := getIndentForLevel(level)
	line := fmt.Sprintf("%s  %s %s", getIndentForLevel(level), icon, step.stepName)
	_ = step.endTime.Day()
	_ = step.startTime.Day()
	duration := step.endTime.Sub(*step.startTime)
	lines := []string{renderLineWithDuration(ctx, line, duration, "", width)}

	if len(step.executions) == 1 {
		// this is a single step pipeline
		// print out the error messages and continue
		for _, se := range step.executions[0].stepErrors {
			lines = append(lines, fmt.Sprintf("%s    └ Error: %s", indent, se))
		}
	}

	// render the executions
	for _, stepExec := range step.executions {
		if stepExec.childPipeline != nil {
			subLines := renderExecutionLog(ctx, stepExec.childPipeline, level+2 /* this is a level higher, since we also need to account for the step line */, width)
			lines = append(lines, subLines...)
		} else if len(stepExec.execKey) != 0 {
			duration := stepExec.endTime.Sub(*step.startTime)
			icon := "🔄"
			if stepExec.status == "failed" {
				icon = "❌"
			}
			eachLine := fmt.Sprintf("%s    %s [%s]", indent, icon, stepExec.execKey)
			lines = append(lines, renderLineWithDuration(ctx, eachLine, duration, "", width))
			for _, se := range stepExec.stepErrors {
				lines = append(lines, fmt.Sprintf("%s      └ Error: %s", indent, se))
			}
		}
	}

	return lines
}

func renderLineWithDuration(ctx context.Context, line string, duration time.Duration, durationPrefix string, width int) string {
	lineWidth := utf8.RuneCountInString(line)

	// HACK: the "⏹️" has 2 code points
	if strings.Contains(line, "⏹️") {
		lineWidth = lineWidth - 2
	}

	durationString := fmt.Sprintf("%s", humanizeDuration(duration))
	if utf8.RuneCountInString(durationPrefix) > 0 {
		durationString = fmt.Sprintf("%s%s", durationPrefix, durationString)
	}
	durationColumnWidth := utf8.RuneCountInString(durationString)

	// {line} {dots} {duration}{durationUnit}
	dotNum := width - (lineWidth + durationColumnWidth + 3 /*accounting for leading and trailing spaces in the dots*/)
	dots := fmt.Sprintf(" %s ", strings.Repeat(".", dotNum))
	rendered := fmt.Sprintf("%s%s%s", line, dots, durationString)

	return rendered
}

func renderPipelineOutput(ctx context.Context, output map[string]any, width int) []string {
	var lines []string
	delete(output, "errors")
	if len(output) >= 1 {
		lines = append(lines, "\nOutput:")
	}
	for k, v := range output {
		line := fmt.Sprintf("➡️ [%s] %v", k, v)
		if utf8.RuneCountInString(line) >= width {
			line = fmt.Sprintf("%s%s", line[0:width-3], "...")
		}
		lines = append(lines, line)
	}

	return lines
}

// humanizeDuration humanizes time.Duration output to a meaningful value,
// golang's default “time.Duration“ output is badly formatted and unreadable.
// TODO: this is salvaged off the internet from https://gist.github.com/harshavardhana/327e0577c4fed9211f65
// with a minor modification for milliseconds. We should be using a library for this
func humanizeDuration(duration time.Duration) string {
	if duration.Milliseconds() < 1000.0 {
		return fmt.Sprintf("%dms", duration.Milliseconds())
	}
	if duration.Seconds() < 60.0 {
		return fmt.Sprintf("%ds", int64(duration.Seconds()))
	}
	if duration.Minutes() < 60.0 {
		remainingSeconds := math.Mod(duration.Seconds(), 60)
		return fmt.Sprintf("%dm%ds", int64(duration.Minutes()), int64(remainingSeconds))
	}
	if duration.Hours() < 24.0 {
		remainingMinutes := math.Mod(duration.Minutes(), 60)
		remainingSeconds := math.Mod(duration.Seconds(), 60)
		return fmt.Sprintf("%dh%dm%ds",
			int64(duration.Hours()), int64(remainingMinutes), int64(remainingSeconds))
	}
	remainingHours := math.Mod(duration.Hours(), 24)
	remainingMinutes := math.Mod(duration.Minutes(), 60)
	remainingSeconds := math.Mod(duration.Seconds(), 60)
	return fmt.Sprintf("%dd%dh%dm%ds", // 00d00h00m00s
		int64(duration.Hours()/24), int64(remainingHours),
		int64(remainingMinutes), int64(remainingSeconds))
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
		icon = "♊"
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
