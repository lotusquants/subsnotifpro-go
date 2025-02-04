package main

import (
	"fmt"
	"log"
	"subsnotifpro-go/config"
	"subsnotifpro-go/database"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to the database (MAKE SURE THIS IS PRESENT)
	database.ConnectDatabase()

	// Run Auto-Migrations
	database.AutoMigrateTables()

	// Initialize Gin router
	router := gin.Default()

	// Health check route
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK", "message": "SubsNotifPro backend is running!"})
	})

	// Start the server with configured port
	log.Printf("Starting server on port %s...\n", cfg.ServerPort)
	router.Run(fmt.Sprintf(":%s", cfg.ServerPort))
}
