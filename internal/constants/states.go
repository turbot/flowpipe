package constants

const (
	StateFailed   = "failed"
	StateFinished = "finished"
	StateSkipped  = "skipped"

	FailureModeIgnored = "ignored"
	FailureModeFailed  = "failed" // rendering output, retry, throw, loop (and other) blocks
)
