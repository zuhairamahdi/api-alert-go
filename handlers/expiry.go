package handlers

import (
	"log"
	"time"

	"api-monitor/database"
)

// CheckExpiredEndpoints checks for expired endpoints and removes them from schedules
func CheckExpiredEndpoints() {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now()
	var expiredIDs []int

	// Find expired endpoints
	for i, endpoint := range endpoints {
		if !endpoint.ExpiresAt.IsZero() && endpoint.ExpiresAt.Before(now) {
			expiredIDs = append(expiredIDs, endpoint.ID)

			// Remove from local endpoints slice
			endpoints = append(endpoints[:i], endpoints[i+1:]...)

			// Delete from database
			if err := database.DeleteEndpoint(endpoint.ID); err != nil {
				log.Printf("Failed to delete expired endpoint %d: %v", endpoint.ID, err)
			} else {
				log.Printf("Deleted expired endpoint: %s (ID: %d)", endpoint.URL, endpoint.ID)
			}
		}
	}

	// Remove expired endpoints from schedules
	if len(expiredIDs) > 0 {
		removeEndpointsFromSchedules(expiredIDs)
	}
}

// removeEndpointsFromSchedules removes the given endpoint IDs from all schedules
func removeEndpointsFromSchedules(endpointIDs []int) {
	for i, schedule := range schedules {
		modified := false
		newEndpoints := make([]int, 0, len(schedule.Endpoints))

		// Filter out expired endpoints
		for _, id := range schedule.Endpoints {
			isExpired := false
			for _, expiredID := range endpointIDs {
				if id == expiredID {
					isExpired = true
					modified = true
					break
				}
			}

			if !isExpired {
				newEndpoints = append(newEndpoints, id)
			}
		}

		// Update schedule if modified
		if modified {
			schedules[i].Endpoints = newEndpoints
			log.Printf("Updated schedule %s (ID: %d) to remove expired endpoints", schedule.Name, schedule.ID)
		}
	}
}
