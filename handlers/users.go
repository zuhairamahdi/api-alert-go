package handlers

import (
	"net/http"
	"strings"
	"time"

	"api-monitor/database"

	"log"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser handles user registration
func CreateUser(c echo.Context) error {
	type UserRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}

	req := new(UserRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
		})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to process password",
		})
	}

	user := &database.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		Name:     req.Name,
		IsActive: true,
	}

	if err := database.DB.Create(user).Error; err != nil {
		if strings.Contains(err.Error(), "uni_users_email") {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "Email already registered",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create user",
		})
	}

	// Create default subscription
	subscription := &database.Subscription{
		UserID:           user.ID,
		PlanName:         "Free",
		MaxEndpoints:     5,
		AllowedIntervals: []int64{5, 60, 300, 900, 1800}, // 5sec, 1min, 5min, 15min, 30min in seconds
		IsActive:         true,
		ExpiresAt:        time.Now().AddDate(0, 1, 0), // 1 month trial
	}

	if err := database.DB.Create(subscription).Error; err != nil {
		log.Printf("Failed to create subscription: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create subscription",
		})
	}

	return c.JSON(http.StatusCreated, user)
}

// GetUser retrieves user information
func GetUser(c echo.Context) error {
	userID := c.Get("user_id").(uint)

	var user database.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "User not found",
		})
	}

	return c.JSON(http.StatusOK, user)
}

// UpdateUser handles user profile updates
func UpdateUser(c echo.Context) error {
	userID := c.Get("user_id").(uint)

	type UpdateRequest struct {
		Name     string `json:"name"`
		Password string `json:"password,omitempty"`
	}

	req := new(UpdateRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
		})
	}

	updates := map[string]interface{}{
		"name": req.Name,
	}

	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to process password",
			})
		}
		updates["password"] = string(hashedPassword)
	}

	if err := database.DB.Model(&database.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update user",
		})
	}

	return c.NoContent(http.StatusOK)
}

// GetSubscription retrieves user's subscription information
func GetSubscription(c echo.Context) error {
	userID := c.Get("user_id").(uint)

	var subscription database.Subscription
	if err := database.DB.Where("user_id = ?", userID).First(&subscription).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Subscription not found",
		})
	}

	return c.JSON(http.StatusOK, subscription)
}
