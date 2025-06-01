package models

import (
	"time"
)

// Todo represents a task in our system
// Now includes audit fields through BaseModel
type Todo struct {
	ID          int        `json:"id" db:"id"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	Completed   bool       `json:"completed" db:"completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at" swaggertype:"string" example:"2024-01-15T15:04:05Z"`
	BaseModel              // Embedded audit fields
}

// CreateTodoRequest represents the data needed to create a new todo
// We separate this from the Todo model to control what users can set
type CreateTodoRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=200" example:"Buy groceries"`
	Description string `json:"description" binding:"max=1000" example:"Milk, bread, eggs, and cheese"`
}

// UpdateTodoRequest represents the data that can be updated
// Using pointers allows us to distinguish between "not provided" and "empty"
type UpdateTodoRequest struct {
	Title       *string `json:"title,omitempty" binding:"omitempty,min=1,max=200" example:"Buy groceries and supplies"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000" example:"Milk, bread, eggs, cheese, and cleaning supplies"`
	Completed   *bool   `json:"completed,omitempty" example:"true"`
}

// TodoWithUser includes user information for created_by and updated_by
type TodoWithUser struct {
	ID            int        `json:"id"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Completed     bool       `json:"completed"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	CreatedBy     *int       `json:"created_by,omitempty"`
	UpdatedBy     *int       `json:"updated_by,omitempty"`
	CreatedByUser *UserInfo  `json:"created_by_user,omitempty"`
	UpdatedByUser *UserInfo  `json:"updated_by_user,omitempty"`
}

// UserInfo represents basic user information for display
type UserInfo struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}
