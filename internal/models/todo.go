package models

import (
	"time"
)

// Todo represents a task in our system
// Notice how we use pointer for CompletedAt - this handles NULL values in the database
type Todo struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Completed   bool       `json:"completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateTodoRequest represents the data needed to create a new todo
// We separate this from the Todo model to control what users can set
type CreateTodoRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=200"`
	Description string `json:"description" binding:"max=1000"`
}

// UpdateTodoRequest represents the data that can be updated
// Using pointers allows us to distinguish between "not provided" and "empty"
type UpdateTodoRequest struct {
	Title       *string `json:"title" binding:"omitempty,min=1,max=200"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
	Completed   *bool   `json:"completed"`
}
