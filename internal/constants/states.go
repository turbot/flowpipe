package constants

const (
	StateFailed   = "failed"
	StateFinished = "finished"
	StateSkipped  = "skipped"

	FailureModeUnexpected = "unexpected" // unexpected primitive error
	FailureModeRuntime    = "runtime"    // expected primitive error, i.e. HTTP status code >= 400
	FailureModeEvaluation = "evaluation" // rendering output, retry, throw, loop (and other) blocks
)
