package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swusjask/todo-api/internal/db"
	"github.com/swusjask/todo-api/internal/handlers"
	"github.com/swusjask/todo-api/internal/models"
	"github.com/swusjask/todo-api/internal/repository"
	"github.com/swusjask/todo-api/internal/service"
)

// testDB holds our test database connection
var testDB *sql.DB

// TestMain runs before all tests and sets up the test environment
func TestMain(m *testing.M) {
	// Setup test database
	// In a real project, use a separate test database
	testDatabaseURL := os.Getenv("TEST_DATABASE_URL")
	if testDatabaseURL == "" {
		testDatabaseURL = "postgres://postgres:password@localhost:5432/todos_test?sslmode=disable"
	}

	var err error
	testDB, err = db.Connect(testDatabaseURL)
	if err != nil {
		fmt.Printf("Failed to connect to test database: %v\n", err)
		os.Exit(1)
	}

	// Run migrations
	if err := db.RunMigrations(testDB, "../../migrations"); err != nil {
		fmt.Printf("Failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	testDB.Close()
	os.Exit(code)
}

// setupRouter creates a test router with all dependencies
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	// Wire up dependencies
	todoRepo := repository.NewTodoRepository(testDB)
	todoService := service.NewTodoService(todoRepo)
	todoHandler := handlers.NewTodoHandler(todoService)

	// Create router
	router := gin.New()
	api := router.Group("/api/v1")
	{
		todos := api.Group("/todos")
		{
			todos.POST("", todoHandler.Create)
			todos.GET("", todoHandler.List)
			todos.GET("/:id", todoHandler.Get)
			todos.PUT("/:id", todoHandler.Update)
			todos.DELETE("/:id", todoHandler.Delete)
		}
	}

	return router
}

// cleanupDatabase removes all test data
// This ensures each test starts with a clean slate
func cleanupDatabase(t *testing.T) {
	_, err := testDB.Exec("TRUNCATE TABLE todos RESTART IDENTITY CASCADE")
	require.NoError(t, err, "Failed to cleanup database")
}

// TestCreateTodo tests the entire create flow
func TestCreateTodo(t *testing.T) {
	cleanupDatabase(t)
	router := setupRouter()

	tests := []struct {
		name       string
		payload    interface{}
		wantStatus int
		wantError  bool
	}{
		{
			name: "valid todo",
			payload: map[string]string{
				"title":       "Test Todo",
				"description": "Test Description",
			},
			wantStatus: http.StatusCreated,
			wantError:  false,
		},
		{
			name: "missing title",
			payload: map[string]string{
				"description": "Test Description",
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name: "title too short",
			payload: map[string]string{
				"title": "Hi",
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/api/v1/todos", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Record response
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.wantStatus, w.Code)

			// Check response body
			if !tt.wantError {
				var todo models.Todo
				err := json.Unmarshal(w.Body.Bytes(), &todo)
				require.NoError(t, err)

				// Verify the todo was created correctly
				assert.NotZero(t, todo.ID)
				assert.Equal(t, "Test Todo", todo.Title)
				assert.False(t, todo.Completed)
				assert.Nil(t, todo.CompletedAt)
			}
		})
	}
}

// TestGetTodo tests retrieving a single todo
func TestGetTodo(t *testing.T) {
	cleanupDatabase(t)
	router := setupRouter()

	// Create a test todo first
	todo := createTestTodo(t, "Test Todo", "Test Description")

	tests := []struct {
		name       string
		todoID     string
		wantStatus int
	}{
		{
			name:       "existing todo",
			todoID:     fmt.Sprintf("%d", todo.ID),
			wantStatus: http.StatusOK,
		},
		{
			name:       "non-existent todo",
			todoID:     "9999",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid ID",
			todoID:     "not-a-number",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/todos/"+tt.todoID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// TestUpdateTodo tests the update functionality
func TestUpdateTodo(t *testing.T) {
	cleanupDatabase(t)
	router := setupRouter()

	// Create a test todo
	todo := createTestTodo(t, "Original Title", "Original Description")

	t.Run("update title only", func(t *testing.T) {
		payload := map[string]interface{}{
			"title": "Updated Title",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/todos/%d", todo.ID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var updated models.Todo
		json.Unmarshal(w.Body.Bytes(), &updated)

		// Title should be updated, description should remain unchanged
		assert.Equal(t, "Updated Title", updated.Title)
		assert.Equal(t, "Original Description", updated.Description)
	})

	t.Run("mark as completed", func(t *testing.T) {
		completed := true
		payload := map[string]interface{}{
			"completed": completed,
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/todos/%d", todo.ID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var updated models.Todo
		json.Unmarshal(w.Body.Bytes(), &updated)

		assert.True(t, updated.Completed)
		assert.NotNil(t, updated.CompletedAt)
	})
}

// TestListTodos tests pagination
func TestListTodos(t *testing.T) {
	cleanupDatabase(t)
	router := setupRouter()

	// Create multiple test todos
	for i := 1; i <= 25; i++ {
		createTestTodo(t, fmt.Sprintf("Todo %d", i), "Description")
	}

	t.Run("default pagination", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/todos", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			Data       []models.Todo `json:"data"`
			Pagination struct {
				Page       int `json:"page"`
				PageSize   int `json:"page_size"`
				TotalCount int `json:"total_count"`
				TotalPages int `json:"total_pages"`
			} `json:"pagination"`
		}

		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Len(t, response.Data, 20) // Default page size
		assert.Equal(t, 1, response.Pagination.Page)
		assert.Equal(t, 25, response.Pagination.TotalCount)
		assert.Equal(t, 2, response.Pagination.TotalPages)
	})

	t.Run("custom page size", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/todos?page=2&page_size=10", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response struct {
			Data []models.Todo `json:"data"`
		}

		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Len(t, response.Data, 10)
	})
}

// Helper function to create a test todo
func createTestTodo(t *testing.T, title, description string) *models.Todo {
	repo := repository.NewTodoRepository(testDB)
	todo, err := repo.Create(context.Background(), &models.CreateTodoRequest{
		Title:       title,
		Description: description,
	})
	require.NoError(t, err)
	return todo
}

// TestDatabaseTransaction tests that our repository handles transactions correctly
func TestDatabaseTransaction(t *testing.T) {
	cleanupDatabase(t)

	// This test demonstrates how you might test transaction handling
	// In a real app, you'd have methods that use transactions

	ctx := context.Background()
	tx, err := testDB.BeginTx(ctx, nil)
	require.NoError(t, err)

	// Try to insert a todo within a transaction
	_, err = tx.ExecContext(ctx,
		"INSERT INTO todos (title, description) VALUES ($1, $2)",
		"Transaction Test", "Testing transactions",
	)
	require.NoError(t, err)

	// Rollback the transaction
	err = tx.Rollback()
	require.NoError(t, err)

	// Verify the todo was not persisted
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM todos WHERE title = $1", "Transaction Test").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "Todo should not exist after rollback")
}
