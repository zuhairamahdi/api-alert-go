package handlers

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"api-monitor/database"
	"api-monitor/models"

	"fmt"

	"log"

	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

var (
	endpoints []models.Endpoint
	mu        sync.RWMutex
)

// LoadEndpoints loads endpoints from the database
func LoadEndpoints() error {
	dbEndpoints, err := database.LoadEndpoints()
	if err != nil {
		return err
	}

	mu.Lock()
	endpoints = dbEndpoints
	mu.Unlock()

	return nil
}

// CreateEndpoint handles the creation of a new endpoint
func CreateEndpoint(c echo.Context) error {
	userID := c.Get("user_id").(uint)

	// Check user's subscription
	var subscription database.Subscription
	if err := database.DB.Where("user_id = ?", userID).First(&subscription).Error; err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "No active subscription found",
		})
	}

	// Check if subscription is active
	if !subscription.IsActive {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "Subscription is not active",
		})
	}

	// Check if user has reached endpoint limit
	var endpointCount int64
	if err := database.DB.Model(&database.Endpoint{}).Where("user_id = ?", userID).Count(&endpointCount).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to check endpoint count",
		})
	}

	if int(endpointCount) >= subscription.MaxEndpoints {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "Endpoint limit reached for your subscription",
		})
	}

	// Bind request to Endpoint model
	endpoint := new(models.Endpoint)
	if err := c.Bind(endpoint); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
		})
	}

	// Validate interval
	isValidInterval := false
	for _, allowedInterval := range subscription.AllowedIntervals {
		if int64(endpoint.Interval) == allowedInterval {
			isValidInterval = true
			break
		}
	}

	if !isValidInterval {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid interval for your subscription",
		})
	}

	// Set default expiry date if not provided (30 days from now)
	if endpoint.ExpiresAt.IsZero() {
		endpoint.ExpiresAt = time.Now().AddDate(0, 0, 30)
	}

	// Create endpoint in database
	dbEndpoint := database.FromModel(*endpoint)
	dbEndpoint.UserID = userID
	dbEndpoint.LastChecked = time.Now()

	if err := database.DB.Create(&dbEndpoint).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create endpoint",
		})
	}

	// Set the endpoint ID from the database
	endpoint.ID = int(dbEndpoint.ID)

	// Create new schedule
	schedule := &database.Schedule{
		Name:      fmt.Sprintf("Schedule for %d second interval", endpoint.Interval),
		Interval:  endpoint.Interval,
		CreatedAt: time.Now(),
		Endpoints: pq.Int64Array{int64(dbEndpoint.ID)},
	}

	// Try to find existing schedule for this interval
	var existingSchedule database.Schedule
	if err := database.DB.Unscoped().Where("interval = ?", endpoint.Interval).First(&existingSchedule).Error; err == nil {
		// Schedule exists, append the new endpoint ID
		existingSchedule.Endpoints = append(existingSchedule.Endpoints, int64(dbEndpoint.ID))
		if err := database.DB.Unscoped().Save(&existingSchedule).Error; err != nil {
			// If schedule update fails, rollback endpoint creation
			database.DB.Delete(&dbEndpoint)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to update schedule",
			})
		}
	} else {
		// Create new schedule
		if err := database.DB.Create(schedule).Error; err != nil {
			// If schedule creation fails, rollback endpoint creation
			database.DB.Delete(&dbEndpoint)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to create schedule",
			})
		}
	}

	// Perform initial health check
	go checkEndpoint(endpoint)

	return c.JSON(http.StatusCreated, endpoint)
}

// GetEndpoints returns all endpoints for the current user
func GetEndpoints(c echo.Context) error {
	userID := c.Get("user_id").(uint)

	var dbEndpoints []database.Endpoint
	if err := database.DB.Where("user_id = ?", userID).Find(&dbEndpoints).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch endpoints",
		})
	}

	// Convert to models
	var userEndpoints []models.Endpoint
	for _, e := range dbEndpoints {
		userEndpoints = append(userEndpoints, e.ToModel())
	}

	return c.JSON(http.StatusOK, userEndpoints)
}

// GetEndpoint returns a specific endpoint by ID
func GetEndpoint(c echo.Context) error {
	userID := c.Get("user_id").(uint)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid ID format",
		})
	}

	var dbEndpoint database.Endpoint
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&dbEndpoint).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Endpoint not found",
		})
	}

	return c.JSON(http.StatusOK, dbEndpoint.ToModel())
}

// UpdateEndpoint updates an existing endpoint
func UpdateEndpoint(c echo.Context) error {
	userID := c.Get("user_id").(uint)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid ID format",
		})
	}

	// Check if endpoint exists and belongs to user
	var existingEndpoint database.Endpoint
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&existingEndpoint).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Endpoint not found",
		})
	}

	// Check subscription for interval validation
	var subscription database.Subscription
	if err := database.DB.Where("user_id = ?", userID).First(&subscription).Error; err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "No active subscription found",
		})
	}

	endpoint := new(models.Endpoint)
	if err := c.Bind(endpoint); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
		})
	}

	// Validate interval
	isValidInterval := false
	for _, allowedInterval := range subscription.AllowedIntervals {
		if int64(endpoint.Interval) == allowedInterval {
			isValidInterval = true
			break
		}
	}

	if !isValidInterval {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid interval for your subscription",
		})
	}

	// Update endpoint
	updates := map[string]interface{}{
		"url":        endpoint.URL,
		"interval":   endpoint.Interval,
		"expires_at": endpoint.ExpiresAt,
	}

	if err := database.DB.Model(&existingEndpoint).Updates(updates).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update endpoint",
		})
	}

	return c.JSON(http.StatusOK, existingEndpoint.ToModel())
}

// DeleteEndpoint removes an endpoint from monitoring
func DeleteEndpoint(c echo.Context) error {
	userID := c.Get("user_id").(uint)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid ID format",
		})
	}

	// Get the endpoint to find its interval
	var endpoint database.Endpoint
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&endpoint).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Endpoint not found",
		})
	}

	// Find the schedule for this interval
	var schedule database.Schedule
	if err := database.DB.Unscoped().Where("interval = ?", endpoint.Interval).First(&schedule).Error; err == nil {
		// Remove the endpoint ID from the schedule's endpoints array
		newEndpoints := make(pq.Int64Array, 0)
		for _, endpointID := range schedule.Endpoints {
			if endpointID != int64(id) {
				newEndpoints = append(newEndpoints, endpointID)
			}
		}
		schedule.Endpoints = newEndpoints

		// If no endpoints left, delete the schedule
		if len(schedule.Endpoints) == 0 {
			if err := database.DB.Unscoped().Delete(&schedule).Error; err != nil {
				log.Printf("Failed to delete empty schedule: %v", err)
			}
		} else {
			// Update the schedule with the new endpoints list
			if err := database.DB.Unscoped().Save(&schedule).Error; err != nil {
				log.Printf("Failed to update schedule: %v", err)
			}
		}
	}

	// Delete the endpoint
	if err := database.DB.Delete(&endpoint).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete endpoint",
		})
	}

	return c.NoContent(http.StatusNoContent)
}
