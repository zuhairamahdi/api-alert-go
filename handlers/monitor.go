package handlers

import (
	"io"
	"log"
	"net/http"
	"time"

	"api-monitor/models"
)

// HealthCheck represents a health check result
type HealthCheck struct {
	ID         int       `json:"id"`
	EndpointID int       `json:"endpoint_id"`
	Status     int       `json:"status"`
	Response   string    `json:"response"`
	CheckedAt  time.Time `json:"checked_at"`
}

var healthChecks []HealthCheck

// CheckAllEndpoints performs health checks on all registered endpoints
func CheckAllEndpoints() {
	mu.RLock()
	endpointsCopy := make([]models.Endpoint, len(endpoints))
	copy(endpointsCopy, endpoints)
	mu.RUnlock()

	now := time.Now()
	for i := range endpointsCopy {
		// Skip expired endpoints
		if !endpointsCopy[i].ExpiresAt.IsZero() && endpointsCopy[i].ExpiresAt.Before(now) {
			continue
		}

		go checkEndpoint(&endpointsCopy[i])
	}
}

func checkEndpoint(endpoint *models.Endpoint) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(endpoint.URL)
	if err != nil {
		updateEndpointStatus(endpoint, "error", 0, err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		updateEndpointStatus(endpoint, "error", resp.StatusCode, "Failed to read response body")
		return
	}

	updateEndpointStatus(endpoint, "ok", resp.StatusCode, string(body))
}

func updateEndpointStatus(endpoint *models.Endpoint, status string, httpStatus int, response string) {
	mu.Lock()
	defer mu.Unlock()

	// Update endpoint status
	for i, e := range endpoints {
		if e.ID == endpoint.ID {
			endpoints[i].Status = status
			endpoints[i].LastChecked = time.Now()
			break
		}
	}

	// Record health check
	healthCheck := HealthCheck{
		ID:         len(healthChecks) + 1,
		EndpointID: endpoint.ID,
		Status:     httpStatus,
		Response:   response,
		CheckedAt:  time.Now(),
	}
	healthChecks = append(healthChecks, healthCheck)

	log.Printf("Health check for %s: Status=%d, Response=%s", endpoint.URL, httpStatus, response)
}
