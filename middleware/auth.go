package middleware

import (
	"net/http"
	"strings"
	"time"

	"api-monitor/database"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

type jwtCustomClaims struct {
	UserID uint `json:"user_id"`
	jwt.StandardClaims
}

// JWT middleware
func JWT(secret []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Missing authorization header",
				})
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid token format",
				})
			}

			token, err := jwt.ParseWithClaims(tokenString, &jwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
				return secret, nil
			})

			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid token",
				})
			}

			claims, ok := token.Claims.(*jwtCustomClaims)
			if !ok || !token.Valid {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid token claims",
				})
			}

			// Check if user exists and is active
			var user database.User
			if err := database.DB.First(&user, claims.UserID).Error; err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "User not found",
				})
			}

			if !user.IsActive {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "User account is inactive",
				})
			}

			c.Set("user_id", claims.UserID)
			return next(c)
		}
	}
}

// GenerateToken generates a new JWT token for a user
func GenerateToken(userID uint, secret []byte) (string, error) {
	claims := &jwtCustomClaims{
		userID,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // 24 hours
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}
