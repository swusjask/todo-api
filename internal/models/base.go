package models

import (
	"context"
	"database/sql"
	"time"
)

// BaseModel contains common fields for all models
// This provides audit trail functionality across all tables
type BaseModel struct {
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy *int      `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy *int      `json:"updated_by,omitempty" db:"updated_by"`
}

// SetCreatedBy sets the created_by field from context
func (b *BaseModel) SetCreatedBy(ctx context.Context) {
	if userID := GetUserIDFromContext(ctx); userID != nil {
		b.CreatedBy = userID
	}
}

// SetUpdatedBy sets the updated_by field from context
func (b *BaseModel) SetUpdatedBy(ctx context.Context) {
	if userID := GetUserIDFromContext(ctx); userID != nil {
		b.UpdatedBy = userID
	}
}

// BeforeCreate sets audit fields before creating a record
func (b *BaseModel) BeforeCreate(ctx context.Context) {
	now := time.Now()
	b.CreatedAt = now
	b.UpdatedAt = now
	b.SetCreatedBy(ctx)
	b.SetUpdatedBy(ctx)
}

// BeforeUpdate sets audit fields before updating a record
func (b *BaseModel) BeforeUpdate(ctx context.Context) {
	b.UpdatedAt = time.Now()
	b.SetUpdatedBy(ctx)
}

// Context key for user information
type contextKey string

const UserContextKey contextKey = "user"

// UserContext holds user information in context
type UserContext struct {
	ID       int
	Email    string
	Username string
	IsAdmin  bool
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) *int {
	if user, ok := ctx.Value(UserContextKey).(*UserContext); ok && user != nil {
		return &user.ID
	}
	return nil
}

// GetUserFromContext extracts full user context
func GetUserFromContext(ctx context.Context) *UserContext {
	if user, ok := ctx.Value(UserContextKey).(*UserContext); ok {
		return user
	}
	return nil
}

// SetUserInContext sets user information in context
func SetUserInContext(ctx context.Context, user *UserContext) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}

// NullInt64 is a helper to convert *int to sql.NullInt64
func NullInt64(i *int) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: int64(*i), Valid: true}
}

// NullInt64ToPtr converts sql.NullInt64 to *int
func NullInt64ToPtr(n sql.NullInt64) *int {
	if !n.Valid {
		return nil
	}
	i := int(n.Int64)
	return &i
}
