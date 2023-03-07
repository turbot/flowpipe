package utils

import "github.com/rs/xid"

func NewUniqueID() string {
	return xid.New().String()
}

func NewExecutionID() string {
	return "exec_" + NewUniqueID()
}

func NewSessionID() string {
	return "sess_" + NewUniqueID()
}

func NewProcessID() string {
	return "p_" + NewUniqueID()
}
