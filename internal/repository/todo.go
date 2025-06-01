package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

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
func (r *TodoRepository) Create(ctx context.Context, req *models.CreateTodoRequest) (*models.Todo, error) {
	todo := &models.Todo{
		Title:       req.Title,
		Description: req.Description,
		Completed:   false,
	}

	// Set audit fields from context
	todo.BeforeCreate(ctx)

	query := `
		INSERT INTO todos (title, description, completed, created_at, updated_at, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, title, description, completed, completed_at, created_at, updated_at, created_by, updated_by
	`

	err := r.db.QueryRowContext(ctx, query,
		todo.Title,
		todo.Description,
		todo.Completed,
		todo.CreatedAt,
		todo.UpdatedAt,
		models.NullInt64(todo.CreatedBy),
		models.NullInt64(todo.UpdatedBy),
	).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.CompletedAt,
		&todo.CreatedAt,
		&todo.UpdatedAt,
		&todo.CreatedBy,
		&todo.UpdatedBy,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create todo: %w", err)
	}

	return todo, nil
}

// GetByID retrieves a single todo
func (r *TodoRepository) GetByID(ctx context.Context, id int) (*models.Todo, error) {
	query := `
		SELECT id, title, description, completed, completed_at, created_at, updated_at, created_by, updated_by
		FROM todos
		WHERE id = $1
	`

	todo := &models.Todo{}
	var createdBy, updatedBy sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.CompletedAt,
		&todo.CreatedAt,
		&todo.UpdatedAt,
		&createdBy,
		&updatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	todo.CreatedBy = models.NullInt64ToPtr(createdBy)
	todo.UpdatedBy = models.NullInt64ToPtr(updatedBy)

	return todo, nil
}

// GetByIDWithUser retrieves a todo with user information
func (r *TodoRepository) GetByIDWithUser(ctx context.Context, id int) (*models.TodoWithUser, error) {
	query := `
		SELECT 
			t.id, t.title, t.description, t.completed, t.completed_at, 
			t.created_at, t.updated_at, t.created_by, t.updated_by,
			cu.id, cu.username, cu.email,
			uu.id, uu.username, uu.email
		FROM todos t
		LEFT JOIN users cu ON t.created_by = cu.id
		LEFT JOIN users uu ON t.updated_by = uu.id
		WHERE t.id = $1
	`

	var (
		todo                         = &models.TodoWithUser{}
		createdBy, updatedBy         sql.NullInt64
		createdByUser, updatedByUser struct {
			ID       sql.NullInt64
			Username sql.NullString
			Email    sql.NullString
		}
	)

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.CompletedAt,
		&todo.CreatedAt,
		&todo.UpdatedAt,
		&createdBy,
		&updatedBy,
		&createdByUser.ID,
		&createdByUser.Username,
		&createdByUser.Email,
		&updatedByUser.ID,
		&updatedByUser.Username,
		&updatedByUser.Email,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get todo with user: %w", err)
	}

	todo.CreatedBy = models.NullInt64ToPtr(createdBy)
	todo.UpdatedBy = models.NullInt64ToPtr(updatedBy)

	if createdByUser.ID.Valid {
		todo.CreatedByUser = &models.UserInfo{
			ID:       int(createdByUser.ID.Int64),
			Username: createdByUser.Username.String,
			Email:    createdByUser.Email.String,
		}
	}

	if updatedByUser.ID.Valid {
		todo.UpdatedByUser = &models.UserInfo{
			ID:       int(updatedByUser.ID.Int64),
			Username: updatedByUser.Username.String,
			Email:    updatedByUser.Email.String,
		}
	}

	return todo, nil
}

// List retrieves todos with pagination
func (r *TodoRepository) List(ctx context.Context, offset, limit int) ([]*models.Todo, int, error) {
	// Get user from context to potentially filter by user
	// userID := models.GetUserIDFromContext(ctx)

	// Base query - you can modify this to filter by user if needed
	countQuery := "SELECT COUNT(*) FROM todos"
	listQuery := `
		SELECT id, title, description, completed, completed_at, created_at, updated_at, created_by, updated_by
		FROM todos
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	// If you want to filter by user, uncomment these:
	// if userID != nil {
	//     countQuery = "SELECT COUNT(*) FROM todos WHERE created_by = $1"
	//     listQuery = `
	//         SELECT id, title, description, completed, completed_at, created_at, updated_at, created_by, updated_by
	//         FROM todos
	//         WHERE created_by = $3
	//         ORDER BY created_at DESC
	//         LIMIT $1 OFFSET $2
	//     `
	// }

	var totalCount int
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count todos: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, listQuery, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list todos: %w", err)
	}
	defer rows.Close()

	var todos []*models.Todo
	for rows.Next() {
		todo := &models.Todo{}
		var createdBy, updatedBy sql.NullInt64

		err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Description,
			&todo.Completed,
			&todo.CompletedAt,
			&todo.CreatedAt,
			&todo.UpdatedAt,
			&createdBy,
			&updatedBy,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan todo: %w", err)
		}

		todo.CreatedBy = models.NullInt64ToPtr(createdBy)
		todo.UpdatedBy = models.NullInt64ToPtr(updatedBy)
		todos = append(todos, todo)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating todos: %w", err)
	}

	return todos, totalCount, nil
}

// Update modifies an existing todo
func (r *TodoRepository) Update(ctx context.Context, id int, req *models.UpdateTodoRequest) (*models.Todo, error) {
	// First, get the existing todo
	existing, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}

	// Set audit fields
	existing.BeforeUpdate(ctx)

	// Build dynamic UPDATE query
	setClauses := []string{"updated_at = $1", "updated_by = $2"}
	args := []interface{}{existing.UpdatedAt, models.NullInt64(existing.UpdatedBy)}
	argIndex := 3

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

		if *req.Completed {
			setClauses = append(setClauses, fmt.Sprintf("completed_at = $%d", argIndex))
			args = append(args, time.Now())
			argIndex++
		} else {
			setClauses = append(setClauses, "completed_at = NULL")
		}
	}

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE todos
		SET %s
		WHERE id = $%d
		RETURNING id, title, description, completed, completed_at, created_at, updated_at, created_by, updated_by
	`, strings.Join(setClauses, ", "), argIndex)

	todo := &models.Todo{}
	var createdBy, updatedBy sql.NullInt64

	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.CompletedAt,
		&todo.CreatedAt,
		&todo.UpdatedAt,
		&createdBy,
		&updatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update todo: %w", err)
	}

	todo.CreatedBy = models.NullInt64ToPtr(createdBy)
	todo.UpdatedBy = models.NullInt64ToPtr(updatedBy)

	return todo, nil
}

// Delete removes a todo from the database
func (r *TodoRepository) Delete(ctx context.Context, id int) error {
	// Optionally, you can check if the user has permission to delete
	// by verifying created_by matches the current user

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
		return sql.ErrNoRows
	}

	return nil
}

// ListByUser retrieves todos created by a specific user
func (r *TodoRepository) ListByUser(ctx context.Context, userID int, offset, limit int) ([]*models.Todo, int, error) {
	countQuery := "SELECT COUNT(*) FROM todos WHERE created_by = $1"
	var totalCount int
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count user todos: %w", err)
	}

	query := `
		SELECT id, title, description, completed, completed_at, created_at, updated_at, created_by, updated_by
		FROM todos
		WHERE created_by = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list user todos: %w", err)
	}
	defer rows.Close()

	var todos []*models.Todo
	for rows.Next() {
		todo := &models.Todo{}
		var createdBy, updatedBy sql.NullInt64

		err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Description,
			&todo.Completed,
			&todo.CompletedAt,
			&todo.CreatedAt,
			&todo.UpdatedAt,
			&createdBy,
			&updatedBy,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan todo: %w", err)
		}

		todo.CreatedBy = models.NullInt64ToPtr(createdBy)
		todo.UpdatedBy = models.NullInt64ToPtr(updatedBy)
		todos = append(todos, todo)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating todos: %w", err)
	}

	return todos, totalCount, nil
}
