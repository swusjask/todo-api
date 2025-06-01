package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/swusjask/todo-api/internal/auth"
	"github.com/swusjask/todo-api/internal/models"
)

// AuthMiddleware creates an authentication middleware
func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		// Validate the token
		claims, err := jwtManager.ValidateToken(bearerToken[1])
		if err != nil {
			if err == auth.ErrExpiredToken {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			}
			c.Abort()
			return
		}

		// Create user context from claims
		userContext := &models.UserContext{
			ID:       claims.UserID,
			Email:    claims.Email,
			Username: claims.Username,
			IsAdmin:  claims.IsAdmin,
		}

		// Set user context in Gin context
		c.Set("user", userContext)

		// Also set in request context for database operations
		ctx := models.SetUserInContext(c.Request.Context(), userContext)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// RequireAdmin middleware ensures the user is an admin
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
			c.Abort()
			return
		}

		user, ok := userInterface.(*models.UserContext)
		if !ok || !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth middleware sets user context if token is provided, but doesn't require it
func OptionalAuth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			c.Next()
			return
		}

		claims, err := jwtManager.ValidateToken(bearerToken[1])
		if err != nil {
			c.Next()
			return
		}

		userContext := &models.UserContext{
			ID:       claims.UserID,
			Email:    claims.Email,
			Username: claims.Username,
			IsAdmin:  claims.IsAdmin,
		}

		c.Set("user", userContext)
		ctx := models.SetUserInContext(c.Request.Context(), userContext)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// GetCurrentUser helper function to get user from Gin context
func GetCurrentUser(c *gin.Context) (*models.UserContext, bool) {
	userInterface, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	user, ok := userInterface.(*models.UserContext)
	return user, ok
}
