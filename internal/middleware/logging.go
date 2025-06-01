package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger creates a custom logging middleware
// This is more informative than Gin's default logger
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer to measure request duration
		start := time.Now()

		// Store the request path - we'll need it for logging
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request - this calls the next handler in the chain
		c.Next()

		// After request is processed, log the details
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		// Add query string if it exists
		if raw != "" {
			path = path + "?" + raw
		}

		// Color-code status codes for better visibility in logs
		statusColor := getStatusColor(statusCode)
		methodColor := getMethodColor(method)
		resetColor := "\033[0m"

		// Log format: [timestamp] IP | latency | method | path | status
		fmt.Printf("[%s] %s | %13v | %s %s %s | %s %3d %s\n",
			start.Format("2006-01-02 15:04:05"),
			clientIP,
			latency,
			methodColor, method, resetColor,
			path,
			statusColor, statusCode, resetColor,
		)
	}
}

// CORS middleware handles Cross-Origin Resource Sharing
// This allows your API to be called from web browsers
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow requests from any origin in development
		// In production, replace * with your specific frontend domain
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		// Handle preflight requests - browsers send these to check permissions
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// getStatusColor returns ANSI color codes based on status code ranges
func getStatusColor(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "\033[42m" // Green background for success
	case code >= 300 && code < 400:
		return "\033[43m" // Yellow background for redirects
	case code >= 400 && code < 500:
		return "\033[41m" // Red background for client errors
	default:
		return "\033[45m" // Magenta background for server errors
	}
}

// getMethodColor returns ANSI color codes for HTTP methods
func getMethodColor(method string) string {
	switch method {
	case "GET":
		return "\033[34m" // Blue
	case "POST":
		return "\033[32m" // Green
	case "PUT":
		return "\033[33m" // Yellow
	case "DELETE":
		return "\033[31m" // Red
	default:
		return "\033[0m" // Default
	}
}
