package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/swusjask/todo-api/internal/models"
	"github.com/swusjask/todo-api/internal/service"
)

// TodoHandler handles HTTP requests for todos
type TodoHandler struct {
	service *service.TodoService
}

func NewTodoHandler(service *service.TodoService) *TodoHandler {
	return &TodoHandler{service: service}
}

// Create handles POST /todos
// @Summary Create a new todo
// @Description Create a new todo item with title and description
// @Tags todos
// @Accept json
// @Produce json
// @Param todo body models.CreateTodoRequest true "Todo object to create"
// @Success 201 {object} models.Todo "Successfully created todo"
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /todos [post]
func (h *TodoHandler) Create(c *gin.Context) {
	var req models.CreateTodoRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	todo, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create todo"})
		return
	}

	c.JSON(http.StatusCreated, todo)
}

// Get handles GET /todos/:id
// @Summary Get a todo by ID
// @Description Get a single todo item by its ID
// @Tags todos
// @Accept json
// @Produce json
// @Param id path int true "Todo ID"
// @Success 200 {object} models.Todo "Todo found"
// @Failure 400 {object} ErrorResponse "Invalid ID format"
// @Failure 404 {object} ErrorResponse "Todo not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /todos/{id} [get]
func (h *TodoHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid ID format"})
		return
	}

	todo, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrTodoNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Todo not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get todo"})
		return
	}

	c.JSON(http.StatusOK, todo)
}

// List handles GET /todos with pagination
// @Summary List todos
// @Description Get a paginated list of todos
// @Tags todos
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20, max: 100)"
// @Success 200 {object} PaginatedTodosResponse "List of todos with pagination"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /todos [get]
func (h *TodoHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	todos, totalCount, err := h.service.List(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to list todos"})
		return
	}

	totalPages := (totalCount + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, PaginatedTodosResponse{
		Data: todos,
		Pagination: PaginationMeta{
			Page:       page,
			PageSize:   pageSize,
			TotalCount: totalCount,
			TotalPages: totalPages,
		},
	})
}

// Update handles PUT /todos/:id
// @Summary Update a todo
// @Description Update an existing todo's title, description, or completion status
// @Tags todos
// @Accept json
// @Produce json
// @Param id path int true "Todo ID"
// @Param todo body models.UpdateTodoRequest true "Todo fields to update"
// @Success 200 {object} models.Todo "Successfully updated todo"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 404 {object} ErrorResponse "Todo not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /todos/{id} [put]
func (h *TodoHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid ID format"})
		return
	}

	var req models.UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	todo, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrTodoNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Todo not found"})
			return
		}
		if errors.Is(err, service.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update todo"})
		return
	}

	c.JSON(http.StatusOK, todo)
}

// Delete handles DELETE /todos/:id
// @Summary Delete a todo
// @Description Delete a todo by its ID
// @Tags todos
// @Accept json
// @Produce json
// @Param id path int true "Todo ID"
// @Success 204 "Todo successfully deleted"
// @Failure 400 {object} ErrorResponse "Invalid ID format"
// @Failure 404 {object} ErrorResponse "Todo not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /todos/{id} [delete]
func (h *TodoHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid ID format"})
		return
	}

	err = h.service.Delete(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrTodoNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Todo not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete todo"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// Response types for Swagger documentation

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error" example:"Invalid request body"`
	Details string `json:"details,omitempty" example:"title is required"`
}

// PaginatedTodosResponse represents a paginated list of todos
type PaginatedTodosResponse struct {
	Data       []*models.Todo `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

// PaginationMeta contains pagination metadata
type PaginationMeta struct {
	Page       int `json:"page" example:"1"`
	PageSize   int `json:"page_size" example:"20"`
	TotalCount int `json:"total_count" example:"100"`
	TotalPages int `json:"total_pages" example:"5"`
}
