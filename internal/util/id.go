package util

import "github.com/rs/xid"

func NewUniqueId() string {
	return xid.New().String()
}

func NewProcessLogId() string {
	return "pl_" + NewUniqueId()
}

func NewExecutionId() string {
	return "exec_" + NewUniqueId()
}

func NewPipelineExecutionId() string {
	return "pexec_" + NewUniqueId()
}

func NewStepExecutionId() string {
	return "sexec_" + NewUniqueId()
}

func NewSessionId() string {
	return "sess_" + NewUniqueId()
}

func NewProcessId() string {
	return "p_" + NewUniqueId()
}
