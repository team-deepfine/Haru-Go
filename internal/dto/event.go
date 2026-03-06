package dto

import "github.com/daewon/haru/internal/model"

// CreateEventRequest is the request body for creating an event.
type CreateEventRequest struct {
	Title           string   `json:"title" binding:"required"`
	StartAt         string   `json:"startAt" binding:"required"`
	EndAt           string   `json:"endAt" binding:"required"`
	AllDay          bool     `json:"allDay"`
	Timezone        string   `json:"timezone"`
	LocationName    *string  `json:"locationName"`
	LocationAddress *string  `json:"locationAddress"`
	LocationLat     *float64 `json:"locationLat"`
	LocationLng     *float64 `json:"locationLng"`
	ReminderOffsets []int64  `json:"reminderOffsets"`
	Notes           *string  `json:"notes"`
}

// UpdateEventRequest is the same structure as CreateEventRequest.
type UpdateEventRequest = CreateEventRequest

// EventListResponse is the response body for listing events.
type EventListResponse struct {
	Events []model.Event `json:"events"`
	Count  int           `json:"count"`
}
