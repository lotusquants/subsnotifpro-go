# 📌 SubsNotifPro Go Backend **Project Log**

## Internal Developer Guide

### Project Overview
This project is a self-hosted subscription management backend written in Go.  
It will handle subscription lifecycle events from Google Play Store and Apple App Store.

### Development Progress

#### ✅ Step 1: Project Setup (04-02-2025)
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

#### ✅ Step 10: Connected Local Project to GitHub & First Commit
- **Initialized Git** in the project (`git init`).
- **Created a GitHub repository**: `subsnotifpro-go`.
- **Linked local repository to GitHub** using:
  ```sh
  git remote add origin https://github.com/lotusquants/subsnotifpro-go.git

### ✅ **Today's Plan (05-02-2025)**
- Work 1 - implement support for **Google Play Service Account Management** as it is critical for integrating with Google Play API. This involves:

1. **Create a Database Table** (`google_play_service_accounts`) for storing service account metadata.
2. **Auto-Migrate the Table** when the server starts.
3. **Implement API to Upload & Store Service Account JSON** securely in `/secrets/` directory.
4. **Implement API to Validate the Service Account JSON** by making a test call to Google Play API.
5. **Implement API to Fetch Validation Status** of the stored service account.
6. **Implement API to Delete the Existing Service Account** (file & metadata).
7. **Set up Background Job** to periodically check if the service account is still valid.
8. **Update Project Log & Commit Changes**.


#### ✅ **Step 11: Added Database Table for Google Play Service Accounts**
- **Created `google_play_service_accounts` table** to store service account metadata.
- **Fields Included:**
  - `id` (Primary Key)
  - `file_path` (Location of uploaded JSON file)
  - `client_email` (Extracted from JSON for quick reference)
  - `valid_status` (Boolean: Whether the account is valid)
  - `last_validated_at` (Timestamp: Last successful validation)
  - `created_at` & `updated_at`
- **Auto-Migration Enabled** to create this table on startup.
- **Tested Table Creation** by checking schema in PostgreSQL:
  ```sh
  psql -U postgres -d subsnotifpro_db -c "\d google_play_service_accounts"

#### ✅ **Step 12: Created API for Uploading Google Play Service Account**
- **Developed API** to upload Google Play service account JSON files to the server.
- **Endpoint**: `/api/google-play/upload-service-account`.
- The API allows users to upload the service account file, stores it in the `secrets/google_play/` directory, and saves metadata (file name, file path, etc.) in the database.
- **Created Test** for the upload service account API to verify its functionality.
- The test simulates a file upload and checks if:
  - The file is saved correctly in the `secrets/google_play/` directory.
  - The metadata (file name, file path, validated status) is saved in the `google_play_service_accounts` table.
- **Ran the Test** and confirmed that the API works as expected:
  - The file is uploaded successfully.
  - The metadata is stored correctly in the database.
- **Tested the API** manually and automatically.
- **Verified** that:
  - The uploaded file is stored in the correct directory.
  - The database contains the expected metadata for the service account.
- **Confirmed Success**:
  - The API returned the expected response.
  - The database was updated with the correct information.

#### ✅ **Step 13: Added API and Tests for Validating Google Play Service Account**
  ##### **1️⃣ Implemented Service Account Validation API**
  - **Created API endpoint:** `GET /api/google-play/validate-service-account`
  - **Purpose:** Validates the stored Google Play service account JSON file by:
    - Ensuring the file exists.
    - Checking the required fields (`type`, `project_id`, `private_key_id`, `private_key`, `client_email`).
    - Attempting an API call to Google Play’s `Monetization.Subscriptions.List` to verify permissions.
  - **Implemented JSON structure validation** to ensure correct service account format.

  ##### **2️⃣ Test Cases for Service Account Validation**
  - **Tested scenarios:**
    - ✅ Valid JSON file passes validation.
    - ❌ Missing required fields result in an error.
    - ❌ Invalid file format is rejected.
    - ❌ If API call fails, validation fails.
    - **Implemented test in `tests/api_tests/validate_google_play_service_account.go`** to automatically verify validation functionality.

  ##### **3️⃣ Best Practice for Managing Service Account Files**
  - Only **one** active service account JSON file is stored at a time.
  - **Previous files are deleted** upon new uploads.
  - Ensures that validation always runs on the **latest uploaded** service account.

  ##### **4️⃣ Verified API and Test Execution**
  - Successfully validated a **real** Google Play service account JSON.
  - API responses correctly indicate success or failure.
  - All test cases passed (`go test ./tests/api_tests/...`).

  ##### **Committed and pushed changes** to GitHub. ([main 479d2f1] Added Google Play service account management: upload, validation API, and tests)


#### ✅ **Step 14 : Today's Work Summary(06-02-2025)**
1. **Implemented Google Play Service Account Management APIs**
   - Added APIs to:
     - Upload service account JSON file.
     - Validate the service account using Google Play API.
     - Retrieve service account status.
     - Delete the service account.
   - Ensured proper database updates for validation status.
   - Implemented structured error handling and logging.

2. **Added Package Name Management APIs**
   - APIs to **set, update, delete, and retrieve** the package name.
   - Implemented database update logic:
     - If settings entry exists, update the package name.
     - If not, create a new settings entry.

3. **Refactored Code for Clean Architecture**
   - Separated **handlers, services, repositories, and validators** into different files.
   - Moved all database interactions to repository layer.
   - Ensured better error handling and logging.

4. **Fixed File Overwriting Issue**
   - Added a **UUID suffix** to service account filenames to prevent overwriting when re-uploading.
   - Ensured old files are deleted before saving new ones.

5. **Fixed Validation and Retrieval Issues**
   - Ensured `service-account-status` API checks for an existing service account before returning details.
   - Updated logic to prevent returning `404` for valid cases.

6. **Commit Details**
   - **Commit Hash**: `8f1ce466d1874ea8f3cbfca3d090ce225569b683`
   - **Commit Message**: `"Implemented Google Play service account & package name management APIs with clean architecture refactor"`

#### ✅ Step 15: Today's Work Summary ( 07-02-2025 and 08-02-2025)
1. **Implemented Real-Time Developer Notifications (RTDN) Processing**
   - Added **Queue Consumer** for processing RTDN events from RabbitMQ.
   - Implemented **DLQ (Dead Letter Queue) handling** for failed events.
   - Added **automatic retries with exponential backoff & jitter**.

2. **Graceful Shutdown & Resource Cleanup**
   - Ensured **safe shutdown of workers**, RabbitMQ consumers, and event processors.
   - Fixed **duplicate database/RabbitMQ closure logs**.
   - **Prometheus monitoring, RabbitMQ reconnections, and database queries now exit cleanly.**

3. **DLQ Monitoring & Debugging APIs**
   - Implemented `/api/google-play/rtdn/dlq/size` to check DLQ size.
   - Added `/api/google-play/rtdn/dlq/retry` to **retry failed webhook events** from DLQ.

#### ✅ **Step 16  - Refactor: Separate consumer and HTTP server logic into distinct services** - (2025-02-08)

- **Completed** the separation of consumer and HTTP server logic into different `main.go` files.

- Successfully ensured that the consumer logic is now independent, running on separate services without interfering with the HTTP server.
- **Addressed** environment variable issues, ensuring proper loading of `.env` files for both services when run separately in different terminals.
- **Verified** that consumers can be scaled independently using different service clusters for restb api and consumers for better scalability and performance.