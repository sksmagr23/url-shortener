package service

import "net/http"

type StatusError struct {
	Code    int
	Message string
}

func (e StatusError) Error() string {
	return e.Message
}

func (e StatusError) StatusCode() int {
	return e.Code
}

func badRequest(message string) error {
	return StatusError{Code: http.StatusBadRequest, Message: message}
}

func unauthorized(message string) error {
	return StatusError{Code: http.StatusUnauthorized, Message: message}
}
