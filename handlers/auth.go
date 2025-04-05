package handlers

import (
	"net/http"

	"api-monitor/database"
	"api-monitor/middleware"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// Login handles user authentication
func Login(c echo.Context) error {
	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	req := new(LoginRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
		})
	}

	var user database.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid credentials",
		})
	}

	if !user.IsActive {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "Account is inactive",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid credentials",
		})
	}

	// Generate JWT token
	token, err := middleware.GenerateToken(user.ID, []byte("your-secret-key")) // Replace with your secret key
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate token",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"token": token,
	})
}
