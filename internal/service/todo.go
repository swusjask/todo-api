package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/swusjask/todo-api/internal/models"
	"github.com/swusjask/todo-api/internal/repository"
)

// Common errors that the service layer might return
var (
	ErrTodoNotFound = errors.New("todo not found")
	ErrInvalidInput = errors.New("invalid input")
)

// TodoService contains business logic for todo operations
// This layer is where you'd add things like validation, authorization, or complex business rules
type TodoService struct {
	repo *repository.TodoRepository
}

func NewTodoService(repo *repository.TodoRepository) *TodoService {
	return &TodoService{repo: repo}
}

// Create validates and creates a new todo
func (s *TodoService) Create(ctx context.Context, req *models.CreateTodoRequest) (*models.Todo, error) {
	// Business logic validation beyond what Gin binding provides
	if len(req.Title) < 3 {
		return nil, fmt.Errorf("%w: title must be at least 3 characters", ErrInvalidInput)
	}

	// In a real app, you might check user permissions here
	// or enforce business rules like "max 100 todos per user"

	return s.repo.Create(ctx, req)
}

// GetByID retrieves a single todo
func (s *TodoService) GetByID(ctx context.Context, id int) (*models.Todo, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid ID", ErrInvalidInput)
	}

	todo, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if todo == nil {
		return nil, ErrTodoNotFound
	}

	return todo, nil
}

// List retrieves todos with pagination
func (s *TodoService) List(ctx context.Context, page, pageSize int) ([]*models.Todo, int, error) {
	// Validate and set defaults for pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20 // Default page size
	}

	offset := (page - 1) * pageSize
	return s.repo.List(ctx, offset, pageSize)
}

// Update modifies an existing todo
func (s *TodoService) Update(ctx context.Context, id int, req *models.UpdateTodoRequest) (*models.Todo, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid ID", ErrInvalidInput)
	}

	// Validate that at least one field is being updated
	if req.Title == nil && req.Description == nil && req.Completed == nil {
		return nil, fmt.Errorf("%w: no fields to update", ErrInvalidInput)
	}

	todo, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}
	if todo == nil {
		return nil, ErrTodoNotFound
	}

	return todo, nil
}

// Delete removes a todo
func (s *TodoService) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("%w: invalid ID", ErrInvalidInput)
	}

	// In a real app, you might check permissions or archive instead of delete

	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTodoNotFound
		}
		return err
	}

	return nil
}
