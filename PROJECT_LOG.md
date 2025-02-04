# SubsNotifPro Go Backend

## Internal Developer Guide

### Project Overview
This project is a self-hosted subscription management backend written in Go.  
It will handle subscription lifecycle events from Google Play Store and Apple App Store.

### Development Progress

#### ✅ Step 1: Project Setup
- Created the project folder `subsnotifpro-go`
- Initialized Go module (`go mod init subsnotifpro-go`)

#### ✅ Step 2: Install Gin Framework
- Installed Gin (`go get -u github.com/gin-gonic/gin`)
- Verified installation (`cat go.mod | grep gin`)

#### ✅ Step 3: Basic HTTP Server
- Created `main.go` and initialized Gin.
- Added `/health` endpoint to check server status.
- Tested with `curl http://localhost:8080/health`.

#### ✅ Step 4: First API Test (Health Check)
- Created `tests/api_tests/health_test.go` for testing `/health` API.
- Installed `testify` package (`go get -u github.com/stretchr/testify`).
- Ran tests successfully using `go test ./tests/api_tests/...`.
- Verified correct response & status code.

#### ✅ Step 5: Configuration Setup & Dynamic Loading
- **Implemented `.env` for configuration management** instead of hardcoding values.
- **Created `.env`** to store server and database configurations securely.
- **Implemented `config.go`** to dynamically load settings from `.env` and environment variables.
- **Updated `main.go`** to load configurations using `config.go` instead of hardcoded values.
- **Successfully ran the server** using `go run main.go` with environment-based configurations.
- **Re-ran all tests (`go test ./tests/api_tests/...`) and verified that the system works correctly.**

#### ✅ Step 6: Database Connections Setup & Testing
- **Implemented dynamic database connection handling** supporting PostgreSQL, MySQL, and SQLite.
- **Installed necessary GORM drivers (`gorm.io/driver/postgres`, `gorm.io/driver/mysql`, `gorm.io/driver/sqlite`).**
- **Updated `database.go` to support multiple databases with dynamic selection based on `.env` settings.**
- **Successfully tested database connections for PostgreSQL (port 5433), MySQL, and SQLite.**
- **Implemented connection pooling for PostgreSQL and MySQL using GORM's default pooling configurations.**
- **Verified `.env` allows switching databases dynamically without code modifications.**
- **Ran API tests for `/health` on all three databases, all tests passed successfully.**

#### ✅ Step 7: Implemented Database Connection Pooling
- **Moved connection pooling settings to `.env`** for flexibility.
- **Added `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, `DB_CONN_MAX_LIFETIME`** environment variables.
- **Modified `database.go`** to dynamically load pooling settings.
- **Restarted server and verified database pooling is correctly applied.**
- **Re-ran API tests (`go test ./tests/api_tests/...`) and confirmed everything works correctly.**

#### ✅ Step 8: Implemented Auto-Migrations for Database Tables
- **Created `models.go`** to dynamically register all models for cleaner imports.
- **Updated `database.go`** to use `models.AllModels` for auto-migrations.
- **Tested migrations using SQLite, PostgreSQL, and MySQL.**
- **Checked database schemas using SQL queries to ensure correct column creation.**
- **Re-ran server (`go run main.go`) and verified logs confirmed migrations completed successfully.**
- **Validated table creation using:**
  - PostgreSQL: `psql -U postgres -d subsnotifpro_db -c "\dt"`
  - MySQL: `SHOW TABLES FROM subsnotifpro_db;`
  - SQLite: `sqlite3 database.db ".tables"`
- **Ensured migrations run before the server starts, preventing missing table issues.**
- **Re-ran API tests (`go test ./tests/api_tests/...`) and confirmed successful execution.**

#### ✅ Step 9: Added Query Logging & Database Tests
- **Implemented Query Logging Feature**, controlled via `DB_QUERY_LOGGING` environment variable.
- **Updated `database.go`** to enable structured query logging when the feature is turned on.
- **Implemented automated tests for:**
  - **Auto-migrations**
  - **Query logging activation & deactivation**
- **Ran tests to verify query logs appear when enabled:**

## ✅ Step 10: Connected Local Project to GitHub & First Commit
- **Initialized Git** in the project (`git init`).
- **Created a GitHub repository**: `subsnotifpro-go`.
- **Linked local repository to GitHub** using:
  ```sh
  git remote add origin https://github.com/your-username/subsnotifpro-go.git