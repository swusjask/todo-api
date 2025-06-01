package models

import (
	"time"
)

// Todo represents a task in our system
type Todo struct {
	ID          int        `json:"id" example:"1"`
	Title       string     `json:"title" example:"Buy groceries"`
	Description string     `json:"description" example:"Milk, bread, eggs, and cheese"`
	Completed   bool       `json:"completed" example:"false"`
	CompletedAt *time.Time `json:"completed_at,omitempty" swaggertype:"string" example:"2024-01-15T15:04:05Z"`
	CreatedAt   time.Time  `json:"created_at" swaggertype:"string" example:"2024-01-15T10:00:00Z"`
	UpdatedAt   time.Time  `json:"updated_at" swaggertype:"string" example:"2024-01-15T10:00:00Z"`
}

// CreateTodoRequest represents the data needed to create a new todo
type CreateTodoRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=200" example:"Buy groceries"`
	Description string `json:"description" binding:"max=1000" example:"Milk, bread, eggs, and cheese"`
}

// UpdateTodoRequest represents the data that can be updated
type UpdateTodoRequest struct {
	Title       *string `json:"title,omitempty" binding:"omitempty,min=1,max=200" example:"Buy groceries and supplies"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000" example:"Milk, bread, eggs, cheese, and cleaning supplies"`
	Completed   *bool   `json:"completed,omitempty" example:"true"`
}
