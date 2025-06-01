package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/swusjask/todo-api/internal/auth"
	"github.com/swusjask/todo-api/internal/models"
	"github.com/swusjask/todo-api/internal/repository"
)

var (
	ErrInvalidCredentials  = errors.New("invalid username or password")
	ErrUserNotActive       = errors.New("user account is not active")
	ErrEmailExists         = errors.New("email already exists")
	ErrUsernameExists      = errors.New("username already exists")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo        *repository.UserRepository
	jwtManager      *auth.JWTManager
	passwordManager *auth.PasswordManager
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo *repository.UserRepository, jwtManager *auth.JWTManager, passwordManager *auth.PasswordManager) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		jwtManager:      jwtManager,
		passwordManager: passwordManager,
	}
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req *models.CreateUserRequest) (*models.UserResponse, error) {
	// Validate password strength
	if err := auth.ValidatePasswordStrength(req.Password); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Normalize email and username
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Username = strings.ToLower(strings.TrimSpace(req.Username))

	// Check if email already exists
	emailExists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if emailExists {
		return nil, ErrEmailExists
	}

	// Check if username already exists
	usernameExists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if usernameExists {
		return nil, ErrUsernameExists
	}

	// Hash the password
	hashedPassword, err := s.passwordManager.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user model
	user := &models.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: hashedPassword,
		FirstName:    strings.TrimSpace(req.FirstName),
		LastName:     strings.TrimSpace(req.LastName),
		IsActive:     true,
		IsAdmin:      false,
	}

	// Save to database
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user.ToResponse(), nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.TokenResponse, error) {
	// Normalize username (could be email or username)
	req.Username = strings.ToLower(strings.TrimSpace(req.Username))

	// Try to find user by username or email
	var user *models.User
	var err error

	// Check if it's an email
	if strings.Contains(req.Username, "@") {
		user, err = s.userRepo.GetByEmail(ctx, req.Username)
	} else {
		user, err = s.userRepo.GetByUsername(ctx, req.Username)
	}

	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserNotActive
	}

	// Verify password
	if err := s.passwordManager.CheckPassword(req.Password, user.PasswordHash); err != nil {
		if errors.Is(err, auth.ErrInvalidPassword) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Generate tokens
	accessToken, err := s.jwtManager.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, expiresAt, err := s.jwtManager.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Save refresh token to database
	if err := s.userRepo.SaveRefreshToken(ctx, user.ID, refreshToken, expiresAt); err != nil {
		return nil, err
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		// Log error but don't fail the login
		// You might want to use a proper logger here
	}

	return &models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    s.jwtManager.GetAccessTokenDuration(),
	}, nil
}

// RefreshToken generates a new access token using a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, req *models.RefreshTokenRequest) (*models.TokenResponse, error) {
	// Validate refresh token format
	if err := s.jwtManager.ValidateRefreshToken(req.RefreshToken); err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// Get refresh token from database
	refreshToken, err := s.userRepo.GetRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}
	if refreshToken == nil {
		return nil, ErrInvalidRefreshToken
	}

	// Get the user
	user, err := s.userRepo.GetByID(ctx, refreshToken.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.IsActive {
		return nil, ErrInvalidRefreshToken
	}

	// Generate new access token
	accessToken, err := s.jwtManager.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Optionally rotate refresh token for better security
	// Delete old token
	if err := s.userRepo.DeleteRefreshToken(ctx, req.RefreshToken); err != nil {
		// Log error but continue
	}

	// Generate new refresh token
	newRefreshToken, expiresAt, err := s.jwtManager.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Save new refresh token
	if err := s.userRepo.SaveRefreshToken(ctx, user.ID, newRefreshToken, expiresAt); err != nil {
		return nil, err
	}

	return &models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    s.jwtManager.GetAccessTokenDuration(),
	}, nil
}

// Logout invalidates the user's refresh token
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.userRepo.DeleteRefreshToken(ctx, refreshToken)
}

// LogoutAll invalidates all refresh tokens for a user
func (s *AuthService) LogoutAll(ctx context.Context, userID int) error {
	return s.userRepo.DeleteUserRefreshTokens(ctx, userID)
}

// GetCurrentUser retrieves the current user information
func (s *AuthService) GetCurrentUser(ctx context.Context, userID int) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user.ToResponse(), nil
}

// CleanupExpiredTokens removes expired refresh tokens (can be run periodically)
func (s *AuthService) CleanupExpiredTokens(ctx context.Context) error {
	return s.userRepo.DeleteExpiredRefreshTokens(ctx)
}
