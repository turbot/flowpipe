package event

var EndEvents = []string{
	"finished",
	"failed",
	"cancelled",
}

const (
	CommandExecutionQueue    = "command.execution_queue"
	HandlerExecutionQueued   = "handler.execution_queued"
	CommandExecutionStart    = "command.execution_start"
	HandlerExecutionStarted  = "handler.execution_started"
	CommandExecutionPlan     = "command.execution_plan"
	HandlerExecutionPlanned  = "handler.execution_planned"
	CommandExecutionFinish   = "command.execution_finish"
	HandlerExecutionFinished = "handler.execution_finished"
	CommandExecutionFail     = "command.execution_fail"
	HandlerExecutionFailed   = "handler.execution_failed"

	CommandPipelineCancel      = "command.pipeline_cancel"
	HandlerPipelineCancelled   = "handler.pipeline_canceled"
	CommandPipelineFail        = "command.pipeline_fail"
	HandlerPipelineFailed      = "handler.pipeline_failed"
	CommandPipelineFinish      = "command.pipeline_finish"
	HandlerPipelineFinished    = "handler.pipeline_finished"
	CommandPipelineLoad        = "command.pipeline_load"
	HandlerPipelineLoaded      = "handler.pipeline_loaded"
	CommandPipelinePause       = "command.pipeline_pause"
	HandlerPipelinePaused      = "handler.pipeline_paused"
	CommandPipelinePlan        = "command.pipeline_plan"
	HandlerPipelinePlanned     = "handler.pipeline_planned"
	CommandPipelineQueue       = "command.pipeline_queue"
	HandlerPipelineQueued      = "handler.pipeline_queued"
	CommandPipelineResume      = "command.pipeline_resume"
	HandlerPipelineResumed     = "handler.pipeline_resumed"
	CommandPipelineStart       = "command.pipeline_start"
	HandlerPipelineStarted     = "handler.pipeline_started"
	HandlerStepFinished        = "handler.step_finished"
	CommandStepForEachPlan     = "command.step_for_each_plan"
	HandlerStepForEachPlanned  = "handler.step_for_each_planned"
	CommandStepPipelineFinish  = "command.step_pipeline_finish"
	HandlerStepPipelineStarted = "handler.step_pipeline_started"
	CommandStepQueue           = "command.step_queue"
	HandlerStepQueued          = "handler.step_queued"
	CommandStepStart           = "command.step_start"

	CommandTriggerQueue    = "command.trigger_queue"
	HandlerTriggerQueued   = "handler.trigger_queued"
	CommandTriggerStart    = "command.trigger_start"
	HandlerTriggerStarted  = "handler.trigger_started"
	CommandTriggerFail     = "command.trigger_fail"
	HandlerTriggerFailed   = "handler.trigger_failed"
	CommandTriggerFinish   = "command.trigger_finish"
	HandlerTriggerFinished = "handler.trigger_finished"
)
