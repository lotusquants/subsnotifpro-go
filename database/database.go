package database

import (
	"fmt"
	"log"

	"subsnotifpro-go/config"
	"subsnotifpro-go/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// ConnectDatabase initializes the database connection and returns a DB instance
func ConnectDatabase(cfg *config.Config) (*gorm.DB, error) {
	var err error
	var dsn string
	var dialector gorm.Dialector

	// Configure GORM logger
	var gormLogger logger.Interface
	if cfg.Database.QueryLogging {
		gormLogger = logger.Default.LogMode(logger.Info) // Log all queries
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent) // Disable query logging
	}

	// Build DSN based on database type and deployment mode
	switch cfg.Database.Type {
	case "postgres":
		dsn, err = buildPostgresDSN(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to build PostgreSQL DSN: %w", err)
		}
		dialector = postgres.Open(dsn)

	case "mysql":
		dsn, err = buildMySQLDSN(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to build MySQL DSN: %w", err)
		}
		dialector = mysql.Open(dsn)

	case "sqlite":
		dsn = cfg.Database.Name
		if dsn == "" {
			dsn = "subsnotifpro.db"
		}
		dialector = sqlite.Open(dsn)

	default:
		return nil, fmt.Errorf("unsupported database type: %s. Available options: postgres, mysql, sqlite", cfg.Database.Type)
	}

	// Open the database with GORM
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "subsnotifpro_",
			SingularTable: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s database: %w", cfg.Database.Type, err)
	}

	// Configure connection pool
	if err := configureConnectionPool(db, cfg); err != nil {
		return nil, fmt.Errorf("failed to configure connection pool: %w", err)
	}

	log.Printf("🚀 Connected to %s database successfully! (Mode: %s)", cfg.Database.Type, cfg.Database.DeploymentMode)
	return db, nil
}

// AutoMigrateTables automatically creates required tables
func AutoMigrateTables(db *gorm.DB, cfg *config.Config) error {
	err := db.AutoMigrate(models.AllModels...)
	if err != nil {
		return fmt.Errorf("auto-migration failed: %w", err)
	}
	log.Printf("✅ Auto-migration completed successfully for %s database! (Mode: %s)", cfg.Database.Type, cfg.Database.DeploymentMode)
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

// buildPostgresDSN builds a PostgreSQL DSN based on deployment mode
func buildPostgresDSN(cfg *config.Config) (string, error) {
	db := cfg.Database

	switch db.DeploymentMode {
	case config.DatabaseDeploymentModeContainer:
		// Standard PostgreSQL connection for container deployment
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			db.Host, db.Port, db.User, db.Password, db.Name, db.SSLMode), nil

	case config.DatabaseDeploymentModeManaged:
		// Azure Database for PostgreSQL with enhanced security
		dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s",
			db.Host, db.Port, db.User, db.Name, db.Azure.SSLMode)

		// Add password if not using managed identity
		if !db.Azure.UseManagedIdentity {
			dsn += fmt.Sprintf(" password=%s", db.Password)
		}

		// Add Azure-specific parameters
		if db.Azure.ConnectTimeout > 0 {
			dsn += fmt.Sprintf(" connect_timeout=%d", int(db.Azure.ConnectTimeout.Seconds()))
		}

		if db.Azure.SSLRootCert != "" {
			dsn += fmt.Sprintf(" sslrootcert=%s", db.Azure.SSLRootCert)
		}

		return dsn, nil

	case config.DatabaseDeploymentModeExternal:
		// External PostgreSQL database
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			db.Host, db.Port, db.User, db.Password, db.Name, db.SSLMode), nil

	default:
		return "", fmt.Errorf("unsupported deployment mode: %s", db.DeploymentMode)
	}
}

// buildMySQLDSN builds a MySQL DSN based on deployment mode
func buildMySQLDSN(cfg *config.Config) (string, error) {
	db := cfg.Database

	switch db.DeploymentMode {
	case config.DatabaseDeploymentModeContainer, config.DatabaseDeploymentModeExternal:
		// Standard MySQL connection
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			db.User, db.Password, db.Host, db.Port, db.Name), nil

	case config.DatabaseDeploymentModeManaged:
		// Azure Database for MySQL with enhanced security
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			db.User, db.Password, db.Host, db.Port, db.Name)

		// Add SSL parameters for Azure
		if db.Azure.SSLMode != "disable" {
			dsn += "&tls=true"
			if db.Azure.SSLRootCert != "" {
				dsn += fmt.Sprintf("&tls-ca=%s", db.Azure.SSLRootCert)
			}
		}

		return dsn, nil

	default:
		return "", fmt.Errorf("unsupported deployment mode: %s", db.DeploymentMode)
	}
}

// configureConnectionPool configures the database connection pool
func configureConnectionPool(db *gorm.DB, cfg *config.Config) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to access database instance: %w", err)
	}

	// Apply connection pool settings
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	log.Printf("🔧 Connection pool configured: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v",
		cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns, cfg.Database.ConnMaxLifetime)

	return nil
}
