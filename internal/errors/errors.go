package errors

import "net/http"

type Error struct {
	Message string `json:"message"`
	Status  int    `json:"code"`
	Details error  `json:"details,omitempty"`
}

func (e Error) Error() string {
	return e.Message
}

func (e Error) WithDetails(details error) Error {
	e.Details = details
	return e
}

func (e Error) WithMessage(message string) Error {
	e.Message = message
	return e
}

var (
	ErrNotFound = Error{Message: "Resource not found", Status: http.StatusNotFound}
	ErrInternal = Error{Message: "Internal server error", Status: http.StatusInternalServerError}
)
