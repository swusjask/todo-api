package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/swusjask/todo-api/docs" // This is required for swagger
	"github.com/swusjask/todo-api/internal/auth"
	"github.com/swusjask/todo-api/internal/config"
	"github.com/swusjask/todo-api/internal/db"
	"github.com/swusjask/todo-api/internal/handlers"
	"github.com/swusjask/todo-api/internal/middleware"
	"github.com/swusjask/todo-api/internal/repository"
	"github.com/swusjask/todo-api/internal/service"
)

// @title Todo API
// @version 1.0
// @description A simple Todo REST API with CRUD operations and JWT authentication
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Connect to database
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Initialize auth components
	jwtManager := auth.NewJWTManager(
		cfg.JWTSecretKey,
		cfg.JWTAccessTokenExpiry,
		cfg.JWTRefreshTokenExpiry,
	)
	passwordManager := auth.NewPasswordManager(cfg.BcryptCost)

	// Initialize repositories
	userRepo := repository.NewUserRepository(database)
	todoRepo := repository.NewTodoRepository(database)

	// Initialize services
	authService := service.NewAuthService(userRepo, jwtManager, passwordManager)
	todoService := service.NewTodoService(todoRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	todoHandler := handlers.NewTodoHandler(todoService)

	// Setup router with auth middleware
	router := setupRouter(cfg, authHandler, todoHandler, jwtManager)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start periodic cleanup of expired tokens
	go func() {
		ticker := time.NewTicker(24 * time.Hour) // Run once per day
		defer ticker.Stop()

		for range ticker.C {
			if err := authService.CleanupExpiredTokens(context.Background()); err != nil {
				log.Printf("Failed to cleanup expired tokens: %v", err)
			}
		}
	}()

	// Start server
	go func() {
		log.Printf("Starting server on port %s", cfg.Port)
		log.Printf("Swagger docs available at http://localhost:%s/swagger/index.html", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}

func setupRouter(cfg *config.Config, authHandler *handlers.AuthHandler, todoHandler *handlers.TodoHandler, jwtManager *auth.JWTManager) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	// Health check endpoint (public)
	// @Summary Health check
	// @Description Check if the API is running
	// @Tags health
	// @Accept json
	// @Produce json
	// @Success 200 {object} map[string]string
	// @Router /health [get]
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Swagger documentation (public)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	api := router.Group("/api/v1")
	{
		// Authentication routes (public)
		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/register", authHandler.Register)
			authRoutes.POST("/login", authHandler.Login)
			authRoutes.POST("/refresh", authHandler.RefreshToken)

			// Protected auth routes
			authProtected := authRoutes.Group("")
			authProtected.Use(middleware.AuthMiddleware(jwtManager))
			{
				authProtected.POST("/logout", authHandler.Logout)
				authProtected.POST("/logout-all", authHandler.LogoutAll)
				authProtected.GET("/me", authHandler.GetMe)
				authProtected.GET("/health", authHandler.HealthCheck)
			}
		}

		// Todo routes (protected)
		todos := api.Group("/todos")
		todos.Use(middleware.AuthMiddleware(jwtManager))
		{
			todos.POST("", todoHandler.Create)
			todos.GET("", todoHandler.List)
			todos.GET("/:id", todoHandler.Get)
			todos.PUT("/:id", todoHandler.Update)
			todos.DELETE("/:id", todoHandler.Delete)
		}

		// Optional: Public todo endpoints with optional auth
		// This allows viewing todos without login but tracks the user if logged in
		publicTodos := api.Group("/public/todos")
		publicTodos.Use(middleware.OptionalAuth(jwtManager))
		{
			publicTodos.GET("", todoHandler.List)
			publicTodos.GET("/:id", todoHandler.Get)
		}

		// Admin routes (example)
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware(jwtManager))
		admin.Use(middleware.RequireAdmin())
		{
			// Add admin endpoints here
			admin.GET("/stats", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "Admin stats endpoint",
					"note":    "Implement admin statistics here",
				})
			})
		}
	}

	return router
}
