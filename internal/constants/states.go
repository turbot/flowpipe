package constants

const (
	StateFailed   = "failed"
	StateFinished = "finished"
	StateSkipped  = "skipped"

	FailureModeIgnored  = "ignored" // ignored=true
	FailureModeStandard = "normal"  // "normal" failure, retry or ignored=true will be followed
	FailureModeFatal    = "fatal"   // rendering output, retry, throw, loop (and other) blocks
)
