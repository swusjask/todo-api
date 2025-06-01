package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/swusjask/todo-api/internal/middleware"
	"github.com/swusjask/todo-api/internal/models"
	"github.com/swusjask/todo-api/internal/service"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register handles user registration
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param user body models.CreateUserRequest true "User registration details"
// @Success 201 {object} models.UserResponse "User created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request or validation error"
// @Failure 409 {object} ErrorResponse "Email or username already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailExists):
			c.JSON(http.StatusConflict, ErrorResponse{Error: "Email already exists"})
		case errors.Is(err, service.ErrUsernameExists):
			c.JSON(http.StatusConflict, ErrorResponse{Error: "Username already exists"})
		case errors.Is(err, service.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to register user"})
		}
		return
	}

	c.JSON(http.StatusCreated, user)
}

// Login handles user authentication
// @Summary Login user
// @Description Authenticate user and get access token
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.TokenResponse "Login successful"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Invalid credentials"
// @Failure 403 {object} ErrorResponse "User account not active"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	tokens, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid username or password"})
		case errors.Is(err, service.ErrUserNotActive):
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "User account is not active"})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to login"})
		}
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Get a new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param token body models.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} models.TokenResponse "Token refreshed successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Invalid refresh token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	tokens, err := h.authService.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidRefreshToken) {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid refresh token"})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to refresh token"})
		}
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// Logout handles user logout
// @Summary Logout user
// @Description Invalidate refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param token body models.RefreshTokenRequest true "Refresh token to invalidate"
// @Success 200 {object} map[string]string "Logout successful"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req models.RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	if err := h.authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

// LogoutAll handles logout from all devices
// @Summary Logout from all devices
// @Description Invalidate all refresh tokens for the user
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string "Logout successful"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	if err := h.authService.LogoutAll(c.Request.Context(), user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to logout from all devices"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out from all devices"})
}

// GetMe handles getting current user information
// @Summary Get current user
// @Description Get the authenticated user's information
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.UserResponse "User information"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	userResponse, err := h.authService.GetCurrentUser(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get user information"})
		return
	}

	c.JSON(http.StatusOK, userResponse)
}

// HealthCheck is a simple authenticated endpoint for testing
// @Summary Health check for authenticated routes
// @Description Check if authentication is working
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Health check response"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /auth/health [get]
func (h *AuthHandler) HealthCheck(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"user_id": user.ID,
		"message": "Authentication is working",
	})
}
