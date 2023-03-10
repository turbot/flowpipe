package utils

import "github.com/rs/xid"

func NewUniqueID() string {
	return xid.New().String()
}

func NewExecutionID() string {
	return "exec_" + NewUniqueID()
}

func NewPipelineExecutionID() string {
	return "pexec_" + NewUniqueID()
}

func NewStepExecutionID() string {
	return "sexec_" + NewUniqueID()
}

func NewSessionID() string {
	return "sess_" + NewUniqueID()
}

func NewProcessID() string {
	return "p_" + NewUniqueID()
}
