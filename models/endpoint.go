package models

import "time"

// Endpoint represents an API endpoint to monitor
type Endpoint struct {
	ID          int       `json:"id"`
	URL         string    `json:"url"`
	Interval    int       `json:"interval"` // in seconds
	LastChecked time.Time `json:"last_checked"`
	Status      string    `json:"status"`
	ExpiresAt   time.Time `json:"expires_at"` // When the endpoint expires
}
