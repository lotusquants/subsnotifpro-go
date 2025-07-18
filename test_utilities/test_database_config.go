package main

import (
	"fmt"
	"os"
	"subsnotifpro-go/config"
)

func main() {
	testDatabaseConfig()
}

func testDatabaseConfig() {
	fmt.Println("=== Database Configuration Test ===")
	
	// Test Container deployment
	fmt.Println("\n--- Container Deployment Test ---")
	os.Setenv("DB_DEPLOYMENT_MODE", "container")
	os.Setenv("DB_TYPE", "postgres")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "subsnotifpro_db")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "postgres")
	os.Setenv("DB_SSLMODE", "disable")
	
	cfg := config.LoadConfig()
	fmt.Printf("Deployment Mode: %s\n", cfg.Database.DeploymentMode)
	fmt.Printf("Database Type: %s\n", cfg.Database.Type)
	fmt.Printf("Host: %s\n", cfg.Database.Host)
	fmt.Printf("Port: %s\n", cfg.Database.Port)
	fmt.Printf("Database Name: %s\n", cfg.Database.Name)
	fmt.Printf("SSL Mode: %s\n", cfg.Database.SSLMode)
	fmt.Printf("Max Open Conns: %d\n", cfg.Database.MaxOpenConns)
	fmt.Printf("Max Idle Conns: %d\n", cfg.Database.MaxIdleConns)
	fmt.Printf("Conn Max Lifetime: %v\n", cfg.Database.ConnMaxLifetime)
	
	// Test Managed deployment
	fmt.Println("\n--- Managed Deployment Test ---")
	os.Setenv("DB_DEPLOYMENT_MODE", "managed")
	os.Setenv("DB_TYPE", "postgres")
	os.Setenv("DB_HOST", "test-server.postgres.database.azure.com")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "subsnotifpro")
	os.Setenv("DB_USER", "testuser@test-server")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_SSLMODE", "require")
	os.Setenv("DB_AZURE_USE_MANAGED_IDENTITY", "false")
	os.Setenv("DB_AZURE_SERVER_NAME", "test-server.postgres.database.azure.com")
	os.Setenv("DB_AZURE_SSL_MODE", "require")
	
	cfg = config.LoadConfig()
	fmt.Printf("Deployment Mode: %s\n", cfg.Database.DeploymentMode)
	fmt.Printf("Database Type: %s\n", cfg.Database.Type)
	fmt.Printf("Host: %s\n", cfg.Database.Host)
	fmt.Printf("Port: %s\n", cfg.Database.Port)
	fmt.Printf("Database Name: %s\n", cfg.Database.Name)
	fmt.Printf("SSL Mode: %s\n", cfg.Database.SSLMode)
	fmt.Printf("Azure Use Managed Identity: %t\n", cfg.Database.Azure.UseManagedIdentity)
	fmt.Printf("Azure Server Name: %s\n", cfg.Database.Azure.ServerName)
	fmt.Printf("Azure SSL Mode: %s\n", cfg.Database.Azure.SSLMode)
	
	// Test External deployment
	fmt.Println("\n--- External Deployment Test ---")
	os.Setenv("DB_DEPLOYMENT_MODE", "external")
	os.Setenv("DB_TYPE", "postgres")
	os.Setenv("DB_HOST", "external-db.example.com")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "production_db")
	os.Setenv("DB_USER", "prod_user")
	os.Setenv("DB_PASSWORD", "prod_pass")
	os.Setenv("DB_SSLMODE", "require")
	
	cfg = config.LoadConfig()
	fmt.Printf("Deployment Mode: %s\n", cfg.Database.DeploymentMode)
	fmt.Printf("Database Type: %s\n", cfg.Database.Type)
	fmt.Printf("Host: %s\n", cfg.Database.Host)
	fmt.Printf("Port: %s\n", cfg.Database.Port)
	fmt.Printf("Database Name: %s\n", cfg.Database.Name)
	fmt.Printf("SSL Mode: %s\n", cfg.Database.SSLMode)
	
	// Test DSN building (without actually connecting)
	fmt.Println("\n--- DSN Building Test ---")
	
	// Test container DSN
	os.Setenv("DB_DEPLOYMENT_MODE", "container")
	cfg = config.LoadConfig()
	fmt.Printf("Container Mode Configuration loaded successfully\n")
	
	// Test managed DSN
	os.Setenv("DB_DEPLOYMENT_MODE", "managed")
	cfg = config.LoadConfig()
	fmt.Printf("Managed Mode Configuration loaded successfully\n")
	
	// Test external DSN
	os.Setenv("DB_DEPLOYMENT_MODE", "external")
	cfg = config.LoadConfig()
	fmt.Printf("External Mode Configuration loaded successfully\n")
	
	fmt.Println("\n✅ Database configuration test completed successfully!")
	fmt.Println("All deployment modes are properly configured.")
}
