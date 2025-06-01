package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/swusjask/todo-api/internal/models"
)

// TodoRepository handles all database operations for todos
type TodoRepository struct {
	db *sql.DB
}

// NewTodoRepository creates a new repository instance
func NewTodoRepository(db *sql.DB) *TodoRepository {
	return &TodoRepository{db: db}
}

// Create inserts a new todo into the database
// Notice how we use RETURNING to get the generated values in a single query
func (r *TodoRepository) Create(ctx context.Context, req *models.CreateTodoRequest) (*models.Todo, error) {
	query := `
		INSERT INTO todos (title, description, completed, created_at, updated_at)
		VALUES ($1, $2, false, NOW(), NOW())
		RETURNING id, title, description, completed, completed_at, created_at, updated_at
	`

	todo := &models.Todo{}
	err := r.db.QueryRowContext(ctx, query, req.Title, req.Description).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.CompletedAt,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create todo: %w", err)
	}

	return todo, nil
}

// GetByID retrieves a single todo
func (r *TodoRepository) GetByID(ctx context.Context, id int) (*models.Todo, error) {
	query := `
		SELECT id, title, description, completed, completed_at, created_at, updated_at
		FROM todos
		WHERE id = $1
	`

	todo := &models.Todo{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.CompletedAt,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not found is not an error at this layer
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	return todo, nil
}

// List retrieves todos with pagination
// Pagination prevents loading thousands of records at once
func (r *TodoRepository) List(ctx context.Context, offset, limit int) ([]*models.Todo, int, error) {
	// First, get the total count for pagination metadata
	var totalCount int
	countQuery := "SELECT COUNT(*) FROM todos"
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count todos: %w", err)
	}

	// Then get the actual records
	query := `
		SELECT id, title, description, completed, completed_at, created_at, updated_at
		FROM todos
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list todos: %w", err)
	}
	defer rows.Close()

	var todos []*models.Todo
	for rows.Next() {
		todo := &models.Todo{}
		err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Description,
			&todo.Completed,
			&todo.CompletedAt,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan todo: %w", err)
		}
		todos = append(todos, todo)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating todos: %w", err)
	}

	return todos, totalCount, nil
}

// Update modifies an existing todo
// This handles partial updates - only updates fields that were provided
func (r *TodoRepository) Update(ctx context.Context, id int, req *models.UpdateTodoRequest) (*models.Todo, error) {
	// Build dynamic UPDATE query based on provided fields
	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIndex := 1

	if req.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argIndex))
		args = append(args, *req.Title)
		argIndex++
	}

	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *req.Description)
		argIndex++
	}

	if req.Completed != nil {
		setClauses = append(setClauses, fmt.Sprintf("completed = $%d", argIndex))
		args = append(args, *req.Completed)
		argIndex++

		// Handle completed_at timestamp
		if *req.Completed {
			setClauses = append(setClauses, "completed_at = NOW()")
		} else {
			setClauses = append(setClauses, "completed_at = NULL")
		}
	}

	// Add ID as the last argument
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE todos
		SET %s
		WHERE id = $%d
		RETURNING id, title, description, completed, completed_at, created_at, updated_at
	`, strings.Join(setClauses, ", "), argIndex)

	todo := &models.Todo{}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.CompletedAt,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update todo: %w", err)
	}

	return todo, nil
}

// Delete removes a todo from the database
func (r *TodoRepository) Delete(ctx context.Context, id int) error {
	query := "DELETE FROM todos WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows // Todo not found
	}

	return nil
}
