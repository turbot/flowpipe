package fperr

import (
	"github.com/go-playground/validator/v10"
)

// As per RFC7807 (https://tools.ietf.org/html/rfc7807) define a standard error model with a limited set of Flowpipe-specific extensions
// Initial inspiration taken from https://github.com/danielgtaylor/huma/blob/master/error.go
type ErrorDetailModel struct {
	Message  string `json:"message"`
	Location string `json:"location,omitempty"`
}

type ErrorModel struct {
	Instance         string              `json:"instance" binding:"required"`
	ID               string              `json:"-"`
	Type             string              `json:"type" binding:"required"`
	Title            string              `json:"title" binding:"required"`
	Status           int                 `json:"status" binding:"required"`
	Detail           string              `json:"detail,omitempty"`
	ValidationErrors []*ErrorDetailModel `json:"validation_errors,omitempty"`

	// All errors are fatal unless specified
	Retryable bool `json:"retryable,omitempty"`
}

func (e ErrorModel) Error() string {
	if e.Detail != "" {
		return e.Title + ": " + e.Detail
	}
	return e.Title
}

func (e ErrorModel) GetStatus() int {
	return e.Status
}

type ValidationError struct {
	Type   string                     `json:"type"`   // Denotes the location where the validation error was encountered.
	Errors validator.ValidationErrors `json:"errors"` // The list of validation errors.
}

func (e ValidationError) Error() string {
	return e.Type + ": " + e.Errors.Error()
}
