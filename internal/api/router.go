package api

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func SetupRouter(handlers *Handlers) *gin.Engine {
	// Create Gin router with default middleware
	r := gin.Default()

	// Initialize security configuration
	securityConfig := NewSecurityConfig()

	// Add security middleware
	r.Use(SecurityHeadersMiddleware())
	r.Use(LoggingMiddleware())
	r.Use(RateLimitMiddleware(securityConfig))
	r.Use(TrustedProxyMiddleware(securityConfig.TrustedProxies))
	r.Use(corsMiddleware())
	r.Use(AuthMiddleware(securityConfig))

	// Health check endpoint
	r.GET("/health", handlers.HealthCheck)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Tag Standards routes
		standards := v1.Group("/standards")
		{
			standards.POST("", handlers.CreateTagStandard)
			standards.POST("/generate", handlers.GenerateTagStandard)
			standards.POST("/validate", handlers.ValidateTagStandardContent)
			standards.GET("", handlers.ListTagStandards)
			standards.GET("/:id", handlers.GetTagStandard)
			standards.PUT("/:id", handlers.UpdateTagStandard)
			standards.DELETE("/:id", handlers.DeleteTagStandard)
		}

		// Operations routes
		operations := v1.Group("/operations")
		{
			operations.POST("", handlers.CreateOperation)
			operations.GET("", handlers.ListOperations)
			operations.GET("/:id", handlers.GetOperation)
			operations.GET("/:id/summary", handlers.GetOperationSummary)
			operations.GET("/:id/results", handlers.GetOperationResults)
			operations.GET("/:id/logs", handlers.GetOperationLogs)
			operations.POST("/:id/execute", handlers.ExecuteOperation)
			operations.POST("/:id/retry", handlers.RetryOperation)
			operations.DELETE("/:id", handlers.DeleteOperation)
		}

		// File explorer routes
		files := v1.Group("/files")
		{
			files.GET("/browse", handlers.BrowseDirectory)
			files.GET("/info", handlers.GetDirectoryInfo)
		}
	}

	// Determine UI build path (for Docker vs local development)
	uiBuildPath := "./web/ui/build"
	if _, err := os.Stat("/usr/share/terratag/web/ui/build"); err == nil {
		uiBuildPath = "/usr/share/terratag/web/ui/build"
	}
	
	// Serve static files for UI
	r.Static("/static", uiBuildPath+"/static")
	r.StaticFile("/", uiBuildPath+"/index.html")
	r.StaticFile("/favicon.svg", uiBuildPath+"/favicon.svg")
	
	// Catch-all route for React Router
	r.NoRoute(func(c *gin.Context) {
		// If it's an API route, return 404
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
			return
		}
		// Otherwise serve the React app
		c.File(uiBuildPath + "/index.html")
	})

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}