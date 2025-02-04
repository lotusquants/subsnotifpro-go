package database

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"subsnotifpro-go/models"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database connection
var DB *gorm.DB

// ConnectDatabase initializes the correct database connection
func ConnectDatabase() {
	var err error

	// Load database type
	dbType := os.Getenv("DB_TYPE")

	var dsn string
	var dialector gorm.Dialector

	// Determine log level for GORM (without affecting other logs)
	queryLogging := os.Getenv("DB_QUERY_LOGGING")
	var gormLogger logger.Interface

	if queryLogging == "true" {
		gormLogger = logger.Default.LogMode(logger.Info) // Log all queries
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent) // Disable query logging
	}

	switch dbType {
	case "postgres":
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("DB_SSLMODE"))
		dialector = postgres.Open(dsn)

	case "mysql":
		dbHost := os.Getenv("MYSQL_HOST")
		dbPort := os.Getenv("MYSQL_PORT")
		dbUser := os.Getenv("MYSQL_USER")
		dbPassword := os.Getenv("MYSQL_PASSWORD")
		dbName := os.Getenv("MYSQL_DB")

		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbUser, dbPassword, dbHost, dbPort, dbName)
		dialector = mysql.Open(dsn)

	case "sqlite":
		dsn = os.Getenv("SQLITE_FILE")
		dialector = sqlite.Open(dsn)

	default:
		log.Fatal("❌ Unsupported database type. Available options: postgres, mysql, sqlite")
	}

	// Open the database with GORM query logging enabled/disabled
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger, // This only affects database query logs
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to %s database: %v", dbType, err)
	}

	// Set up connection pooling
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("❌ Failed to access database instance: %v", err)
	}

	// Load pooling configurations from .env
	maxOpenConns, _ := strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONNS"))
	maxIdleConns, _ := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNS"))
	connMaxLifetime, _ := time.ParseDuration(os.Getenv("DB_CONN_MAX_LIFETIME"))

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	log.Printf("🚀 Connected to %s database successfully!", dbType)
}

// AutoMigrateTables automatically creates required tables
func AutoMigrateTables() {
	err := DB.AutoMigrate(models.AllModels...)
	if err != nil {
		log.Fatalf("❌ Auto-migration failed: %v", err)
	}
	log.Println("✅ Auto-migration for test models completed successfully!")
}
