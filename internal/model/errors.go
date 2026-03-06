package model

import "errors"

var (
	ErrEventNotFound    = errors.New("event not found")
	ErrInvalidTimeRange = errors.New("end time must be after start time")
	ErrTitleRequired    = errors.New("title is required")
	ErrInvalidTimezone  = errors.New("invalid timezone")
)
