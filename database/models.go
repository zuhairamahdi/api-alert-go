package database

import (
	"api-monitor/models"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Predefined intervals in seconds
const (
	Interval5Sec  = 5
	Interval1Min  = 60
	Interval5Min  = 300
	Interval15Min = 900
	Interval30Min = 1800
	Interval1Hour = 3600
)

// User represents a system user
type User struct {
	gorm.Model
	Email     string     `json:"email" gorm:"unique"`
	Password  string     `json:"-"` // Password hash, not exposed in JSON
	Name      string     `json:"name"`
	IsActive  bool       `json:"is_active" gorm:"default:true"`
	Endpoints []Endpoint `json:"endpoints" gorm:"foreignKey:UserID"`
}

// Subscription represents a user's subscription plan
type Subscription struct {
	gorm.Model
	UserID           uint          `json:"user_id"`
	PlanName         string        `json:"plan_name"`
	MaxEndpoints     int           `json:"max_endpoints"`
	AllowedIntervals pq.Int64Array `json:"allowed_intervals" gorm:"type:integer[]"`
	IsActive         bool          `json:"is_active" gorm:"default:true"`
	ExpiresAt        time.Time     `json:"expires_at"`
}

// Endpoint represents an API endpoint to monitor
type Endpoint struct {
	gorm.Model
	UserID      uint      `json:"user_id"`
	URL         string    `json:"url"`
	Interval    int       `json:"interval"` // in seconds
	LastChecked time.Time `json:"last_checked"`
	Status      string    `json:"status"`
	ExpiresAt   time.Time `json:"expires_at"` // When the endpoint expires
}

// ToModel converts a database Endpoint to a models.Endpoint
func (e *Endpoint) ToModel() models.Endpoint {
	return models.Endpoint{
		ID:          int(e.ID),
		URL:         e.URL,
		Interval:    e.Interval,
		LastChecked: e.LastChecked,
		Status:      e.Status,
		ExpiresAt:   e.ExpiresAt,
	}
}

// FromModel creates a database Endpoint from a models.Endpoint
func FromModel(e models.Endpoint) Endpoint {
	return Endpoint{
		URL:         e.URL,
		Interval:    e.Interval,
		LastChecked: e.LastChecked,
		Status:      e.Status,
		ExpiresAt:   e.ExpiresAt,
	}
}

// Schedule represents a monitoring schedule
type Schedule struct {
	gorm.Model
	Name      string        `json:"name"`
	Interval  int           `json:"interval"` // in seconds
	CreatedAt time.Time     `json:"created_at"`
	Endpoints pq.Int64Array `json:"endpoints" gorm:"type:integer[]"` // List of endpoint IDs
}

// HealthCheck represents a health check result
type HealthCheck struct {
	gorm.Model
	EndpointID int       `json:"endpoint_id"`
	Status     int       `json:"status"`
	Response   string    `json:"response"`
	CheckedAt  time.Time `json:"checked_at"`
}
