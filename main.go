package main

import (
	"fmt"
	"log"
	"subsnotifpro-go/config"
	"subsnotifpro-go/database"
	"subsnotifpro-go/routes"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to the database
	database.ConnectDatabase()

	// Run Auto-Migrations
	database.AutoMigrateTables()

	// Initialize router
	router := routes.SetupRouter()

	// Start the server with the configured port
	log.Printf("🚀 Starting server on port %s...\n", cfg.ServerPort)
	router.Run(fmt.Sprintf(":%s", cfg.ServerPort))
}
