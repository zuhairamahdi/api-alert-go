package handlers

import "errors"

var (
	ErrEndpointNotFound = errors.New("endpoint not found")
	ErrScheduleNotFound = errors.New("schedule not found")
)
