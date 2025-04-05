package handlers

import (
	"log"
	"sync"
	"time"

	"api-monitor/models"
)

// Schedule represents a monitoring schedule
type Schedule struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Interval  int       `json:"interval"` // in seconds
	CreatedAt time.Time `json:"created_at"`
	Endpoints []int     `json:"endpoints"` // List of endpoint IDs
}

var (
	schedules        []Schedule
	scheduleMu       sync.RWMutex
	runningSchedules map[int]*time.Ticker
)

// CreateSchedule creates a new monitoring schedule
func CreateSchedule(name string, interval int, endpointIDs []int) (*Schedule, error) {
	scheduleMu.Lock()
	defer scheduleMu.Unlock()

	// Validate endpoint IDs exist
	for _, id := range endpointIDs {
		found := false
		for _, endpoint := range endpoints {
			if endpoint.ID == id {
				found = true
				break
			}
		}
		if !found {
			return nil, ErrEndpointNotFound
		}
	}

	schedule := Schedule{
		ID:        len(schedules) + 1,
		Name:      name,
		Interval:  interval,
		CreatedAt: time.Now(),
		Endpoints: endpointIDs,
	}

	schedules = append(schedules, schedule)

	// Start the schedule
	startSchedule(schedule)

	return &schedule, nil
}

// GetSchedules returns all monitoring schedules
func GetSchedules() []Schedule {
	scheduleMu.RLock()
	defer scheduleMu.RUnlock()
	return schedules
}

// GetSchedule returns a specific schedule by ID
func GetSchedule(id int) (*Schedule, error) {
	scheduleMu.RLock()
	defer scheduleMu.RUnlock()

	for _, schedule := range schedules {
		if schedule.ID == id {
			return &schedule, nil
		}
	}
	return nil, ErrScheduleNotFound
}

// UpdateSchedule updates an existing schedule
func UpdateSchedule(id int, name string, interval int, endpointIDs []int) (*Schedule, error) {
	scheduleMu.Lock()
	defer scheduleMu.Unlock()

	// Validate endpoint IDs exist
	for _, endpointID := range endpointIDs {
		found := false
		for _, endpoint := range endpoints {
			if endpoint.ID == endpointID {
				found = true
				break
			}
		}
		if !found {
			return nil, ErrEndpointNotFound
		}
	}

	for i, schedule := range schedules {
		if schedule.ID == id {
			// Stop existing schedule
			stopSchedule(id)

			// Update schedule
			schedules[i].Name = name
			schedules[i].Interval = interval
			schedules[i].Endpoints = endpointIDs

			// Start new schedule
			startSchedule(schedules[i])

			return &schedules[i], nil
		}
	}

	return nil, ErrScheduleNotFound
}

// DeleteSchedule removes a schedule
func DeleteSchedule(id int) error {
	scheduleMu.Lock()
	defer scheduleMu.Unlock()

	for i, schedule := range schedules {
		if schedule.ID == id {
			stopSchedule(id)
			schedules = append(schedules[:i], schedules[i+1:]...)
			return nil
		}
	}

	return ErrScheduleNotFound
}

func startSchedule(schedule Schedule) {
	if runningSchedules == nil {
		runningSchedules = make(map[int]*time.Ticker)
	}

	ticker := time.NewTicker(time.Duration(schedule.Interval) * time.Second)
	runningSchedules[schedule.ID] = ticker

	go func() {
		for range ticker.C {
			scheduleMu.RLock()
			// Get current endpoints for this schedule
			var currentEndpoints []models.Endpoint
			for _, endpointID := range schedule.Endpoints {
				for _, endpoint := range endpoints {
					if endpoint.ID == endpointID {
						currentEndpoints = append(currentEndpoints, endpoint)
						break
					}
				}
			}
			scheduleMu.RUnlock()

			// Check all endpoints in this schedule
			for i := range currentEndpoints {
				go checkEndpoint(&currentEndpoints[i])
			}
		}
	}()

	log.Printf("Started schedule %s (ID: %d) with interval %d seconds", schedule.Name, schedule.ID, schedule.Interval)
}

func stopSchedule(id int) {
	if ticker, exists := runningSchedules[id]; exists {
		ticker.Stop()
		delete(runningSchedules, id)
		log.Printf("Stopped schedule ID: %d", id)
	}
}
