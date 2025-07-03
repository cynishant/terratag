package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/cloudyali/terratag/internal/api"
	"github.com/cloudyali/terratag/internal/services"
)

func main() {
	// Configure logrus
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
		PadLevelText:  true,
	})
	logrus.SetLevel(logrus.InfoLevel)
	
	// Get configuration from environment
	dbPath := getEnv("DB_PATH", "./terratag.db")
	port := getEnv("PORT", "8080")
	logLevel := getEnv("LOG_LEVEL", "info")
	
	// Set log level based on environment
	if level, err := logrus.ParseLevel(logLevel); err == nil {
		logrus.SetLevel(level)
	}

	logrus.WithFields(logrus.Fields{
		"component": "main",
		"action":    "startup",
	}).Info("Starting Terratag API Server")

	logrus.WithFields(logrus.Fields{
		"database": dbPath,
		"port":     port,
		"logLevel": logLevel,
	}).Info("Server configuration")

	// Initialize database service
	logrus.WithField("component", "database").Info("Initializing database service")
	dbService, err := services.NewDatabaseService(dbPath)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "database",
			"error":     err.Error(),
			"dbPath":    dbPath,
		}).Fatal("Failed to initialize database service")
	}
	defer dbService.Close()
	logrus.WithField("component", "database").Info("Database service initialized successfully")

	// Initialize services
	logrus.WithField("component", "services").Info("Initializing application services")
	tagStandardsService := services.NewTagStandardsService(dbService)
	operationsService := services.NewOperationsService(dbService, tagStandardsService)
	logrus.WithField("component", "services").Info("Services initialized successfully")

	// Initialize handlers
	handlers := api.NewHandlers(tagStandardsService, operationsService, dbService)

	// Setup router
	router := api.SetupRouter(handlers)

	// Start server
	logrus.WithFields(logrus.Fields{
		"component":    "server",
		"port":         port,
		"apiEndpoint":  "http://localhost:" + port + "/api/v1",
		"ui":          "http://localhost:" + port,
		"healthCheck": "http://localhost:" + port + "/health",
	}).Info("====== Terratag API Server Ready ======")

	if err := router.Run(":" + port); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "server",
			"error":     err.Error(),
			"port":      port,
		}).Fatal("Failed to start server")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}