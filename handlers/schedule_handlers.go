package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// ScheduleRequest represents the request body for creating/updating a schedule
type ScheduleRequest struct {
	Name      string `json:"name"`
	Interval  int    `json:"interval"`
	Endpoints []int  `json:"endpoints"`
}

// CreateScheduleHandler handles the creation of a new schedule
func CreateScheduleHandler(c echo.Context) error {
	req := new(ScheduleRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
		})
	}

	schedule, err := CreateSchedule(req.Name, req.Interval, req.Endpoints)
	if err != nil {
		if err == ErrEndpointNotFound {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "One or more endpoints not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create schedule",
		})
	}

	return c.JSON(http.StatusCreated, schedule)
}

// GetSchedulesHandler returns all schedules
func GetSchedulesHandler(c echo.Context) error {
	schedules := GetSchedules()
	return c.JSON(http.StatusOK, schedules)
}

// GetScheduleHandler returns a specific schedule
func GetScheduleHandler(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid ID format",
		})
	}

	schedule, err := GetSchedule(id)
	if err != nil {
		if err == ErrScheduleNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Schedule not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get schedule",
		})
	}

	return c.JSON(http.StatusOK, schedule)
}

// UpdateScheduleHandler updates an existing schedule
func UpdateScheduleHandler(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid ID format",
		})
	}

	req := new(ScheduleRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
		})
	}

	schedule, err := UpdateSchedule(id, req.Name, req.Interval, req.Endpoints)
	if err != nil {
		if err == ErrScheduleNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Schedule not found",
			})
		}
		if err == ErrEndpointNotFound {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "One or more endpoints not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update schedule",
		})
	}

	return c.JSON(http.StatusOK, schedule)
}

// DeleteScheduleHandler removes a schedule
func DeleteScheduleHandler(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid ID format",
		})
	}

	err = DeleteSchedule(id)
	if err != nil {
		if err == ErrScheduleNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Schedule not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete schedule",
		})
	}

	return c.NoContent(http.StatusNoContent)
}
