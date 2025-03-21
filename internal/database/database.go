package database

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"subsnotifpro-go/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// ConnectDatabase initializes the database connection and returns a DB instance
func ConnectDatabase() (*gorm.DB, error) {
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
		return nil, fmt.Errorf("❌ Unsupported database type. Available options: postgres, mysql, sqlite")
	}

	// Open the database with GORM query logging enabled/disabled with custom naming strategy
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger, // This only affects database query logs
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "subsnotifpro_",
			SingularTable: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("❌ Failed to connect to %s database: %w", dbType, err)
	}

	// Set up connection pooling
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("❌ Failed to access database instance: %w", err)
	}

	// Load pooling configurations from .env
	maxOpenConns, _ := strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONNS"))
	maxIdleConns, _ := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNS"))
	connMaxLifetime, _ := time.ParseDuration(os.Getenv("DB_CONN_MAX_LIFETIME"))

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	log.Printf("🚀 Connected to %s database successfully!", dbType)
	return db, nil
}

// AutoMigrateTables automatically creates required tables
func AutoMigrateTables(db *gorm.DB) error {
	err := db.AutoMigrate(models.AllModels...)
	if err != nil {
		return fmt.Errorf("❌ Auto-migration failed: %w", err)
	}
	log.Printf("✅ Auto-migration completed successfully for %s database!", os.Getenv("DB_TYPE"))
	return nil
}

// ApplyCompositeIndexes manually adds composite indexes for performance tuning.
func ApplyCompositeIndexes(db *gorm.DB) error {
	compositeIndexes := []string{
		// SubscriptionPurchaseV2 composite indexes
		`CREATE INDEX IF NOT EXISTS idx_subscription_user_plan_type ON subsnotifpro_subscription_purchase_v2 (user_id, plan_type)`,
		`CREATE INDEX IF NOT EXISTS idx_subscription_plan_start_time ON subsnotifpro_subscription_purchase_v2 (plan_type, start_time)`,

		// AutoRenewingPlanHistory composite indexes
		`CREATE INDEX IF NOT EXISTS idx_auto_renewing_history_subscription_change ON subsnotifpro_auto_renewing_plan_history (subscription_id, change_type)`,
		`CREATE INDEX IF NOT EXISTS idx_auto_renewing_history_change_time ON subsnotifpro_auto_renewing_plan_history (change_time)`,

		// PrepaidPlanHistory composite indexes
		`CREATE INDEX IF NOT EXISTS idx_prepaid_history_subscription_change ON subsnotifpro_prepaid_plan_history (subscription_id, change_type)`,
		`CREATE INDEX IF NOT EXISTS idx_prepaid_history_change_time ON subsnotifpro_prepaid_plan_history (change_time)`,

		// DeferredItemReplacementHistory composite indexes
		`CREATE INDEX IF NOT EXISTS idx_deferred_item_replacement_subscription ON subsnotifpro_deferred_item_replacement_history (subscription_id, plan_type)`,

		// SignupPromotionHistory composite indexes
		`CREATE INDEX IF NOT EXISTS idx_signup_promotion_subscription ON subsnotifpro_signup_promotion_history (subscription_id, plan_type)`,
	}

	for _, query := range compositeIndexes {
		if err := db.Exec(query).Error; err != nil {
			return err
		}
	}
	return nil
}

func CloseDatabase(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Println("❌ Error getting raw DB instance:", err)
		return
	}

	log.Println("🚦 Closing database connection...")
	sqlDB.Close()
	log.Println("✅ Database connection closed")
}
