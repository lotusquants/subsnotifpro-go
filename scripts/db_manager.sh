#!/bin/bash

# Database Migration and Management Script
# This script provides automated database operations for different deployment modes

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# Load environment variables
load_env() {
    if [ -f "$PROJECT_ROOT/.env" ]; then
        # Process .env file and export variables, handling inline comments
        while IFS='=' read -r key value; do
            # Skip empty lines and comments
            if [[ -z "$key" || "$key" =~ ^#.* ]]; then
                continue
            fi
            # Remove inline comments from value
            value=$(echo "$value" | sed 's/#.*//' | xargs)
            # Export the variable
            export "$key"="$value"
        done < <(grep -v '^#' "$PROJECT_ROOT/.env" | grep -v '^$')
        log_info "Environment variables loaded from .env file"
    else
        log_warn "No .env file found, using system environment variables"
    fi
}

# Get database connection parameters based on deployment mode
get_db_connection() {
    local mode=${DB_DEPLOYMENT_MODE:-"container"}
    
    case $mode in
        "container")
            DB_HOST=${DB_HOST:-"localhost"}
            DB_PORT=${DB_PORT:-"5432"}
            DB_USER=${DB_USER:-"postgres"}
            DB_PASSWORD=${DB_PASSWORD:-"postgres"}
            DB_NAME=${DB_NAME:-"subsnotifpro_db"}
            DB_SSLMODE=${DB_SSLMODE:-"disable"}
            ;;
        "managed")
            DB_HOST=${DB_HOST:-"your-server.postgres.database.azure.com"}
            DB_PORT=${DB_PORT:-"5432"}
            DB_USER=${DB_USER:-"your-user"}
            DB_PASSWORD=${DB_PASSWORD:-"your-password"}
            DB_NAME=${DB_NAME:-"subsnotifpro_db"}
            DB_SSLMODE=${DB_SSLMODE:-"require"}
            ;;
        "external")
            DB_HOST=${DB_HOST:-"localhost"}
            DB_PORT=${DB_PORT:-"5432"}
            DB_USER=${DB_USER:-"postgres"}
            DB_PASSWORD=${DB_PASSWORD:-"postgres"}
            DB_NAME=${DB_NAME:-"subsnotifpro_db"}
            DB_SSLMODE=${DB_SSLMODE:-"disable"}
            ;;
        *)
            log_error "Unknown deployment mode: $mode"
            exit 1
            ;;
    esac
    
    # Construct connection string
    export PGPASSWORD="$DB_PASSWORD"
    DB_CONNECTION_STRING="postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=$DB_SSLMODE"
    
    log_info "Database connection configured for mode: $mode"
    log_debug "Host: $DB_HOST:$DB_PORT, Database: $DB_NAME, SSL: $DB_SSLMODE"
}

# Test database connectivity
test_connection() {
    log_info "Testing database connectivity..."
    
    if psql "$DB_CONNECTION_STRING" -c "SELECT 1;" > /dev/null 2>&1; then
        log_info "Database connection successful"
        return 0
    else
        log_error "Database connection failed"
        return 1
    fi
}

# Wait for database to be ready
wait_for_database() {
    local max_attempts=30
    local attempt=1
    
    log_info "Waiting for database to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        if test_connection; then
            log_info "Database is ready"
            return 0
        fi
        
        log_debug "Attempt $attempt/$max_attempts failed, waiting 2 seconds..."
        sleep 2
        ((attempt++))
    done
    
    log_error "Database did not become ready within timeout"
    return 1
}

# Create database if it doesn't exist
create_database() {
    log_info "Creating database if it doesn't exist..."
    
    # Connect to postgres database to create our database
    local postgres_connection="postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/postgres?sslmode=$DB_SSLMODE"
    
    # Check if database exists
    if psql "$postgres_connection" -lqt | cut -d \| -f 1 | grep -qw "$DB_NAME"; then
        log_info "Database '$DB_NAME' already exists"
    else
        log_info "Creating database '$DB_NAME'..."
        psql "$postgres_connection" -c "CREATE DATABASE $DB_NAME;"
        log_info "Database '$DB_NAME' created successfully"
    fi
}

# Run database migrations
run_migrations() {
    log_info "Running database migrations..."
    
    # Check if migration tool is available
    if ! command -v migrate &> /dev/null; then
        log_error "Migration tool 'migrate' not found. Please install golang-migrate."
        log_info "Install with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
        return 1
    fi
    
    # Check if migrations directory exists
    local migrations_dir="$PROJECT_ROOT/migrations"
    if [ ! -d "$migrations_dir" ]; then
        log_warn "No migrations directory found at $migrations_dir"
        return 0
    fi
    
    # Run migrations
    migrate -path "$migrations_dir" -database "$DB_CONNECTION_STRING" up
    
    if [ $? -eq 0 ]; then
        log_info "Migrations completed successfully"
    else
        log_error "Migration failed"
        return 1
    fi
}

# Create database backup
create_backup() {
    local backup_dir="${1:-$PROJECT_ROOT/backups}"
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_file="$backup_dir/backup_${DB_NAME}_${timestamp}.sql"
    
    log_info "Creating database backup..."
    
    # Create backup directory if it doesn't exist
    mkdir -p "$backup_dir"
    
    # Create backup
    pg_dump "$DB_CONNECTION_STRING" > "$backup_file"
    
    if [ $? -eq 0 ]; then
        log_info "Backup created successfully: $backup_file"
        
        # Compress backup
        gzip "$backup_file"
        log_info "Backup compressed: $backup_file.gz"
        
        # Clean up old backups (keep last 7 days)
        find "$backup_dir" -name "backup_${DB_NAME}_*.sql.gz" -mtime +7 -delete
        log_info "Old backups cleaned up"
        
        echo "$backup_file.gz"
    else
        log_error "Backup failed"
        return 1
    fi
}

# Restore database from backup
restore_backup() {
    local backup_file="$1"
    
    if [ -z "$backup_file" ]; then
        log_error "Backup file not specified"
        return 1
    fi
    
    if [ ! -f "$backup_file" ]; then
        log_error "Backup file not found: $backup_file"
        return 1
    fi
    
    log_info "Restoring database from backup: $backup_file"
    
    # If backup is compressed, decompress it
    if [[ "$backup_file" == *.gz ]]; then
        log_info "Decompressing backup file..."
        gunzip -c "$backup_file" | psql "$DB_CONNECTION_STRING"
    else
        psql "$DB_CONNECTION_STRING" < "$backup_file"
    fi
    
    if [ $? -eq 0 ]; then
        log_info "Database restored successfully"
    else
        log_error "Database restore failed"
        return 1
    fi
}

# Reset database (drop and recreate)
reset_database() {
    log_warn "Resetting database - this will delete all data!"
    
    read -p "Are you sure you want to continue? (y/N): " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Database reset cancelled"
        return 0
    fi
    
    # Connect to postgres database to drop our database
    local postgres_connection="postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/postgres?sslmode=$DB_SSLMODE"
    
    log_info "Dropping database '$DB_NAME'..."
    psql "$postgres_connection" -c "DROP DATABASE IF EXISTS $DB_NAME;"
    
    log_info "Creating database '$DB_NAME'..."
    psql "$postgres_connection" -c "CREATE DATABASE $DB_NAME;"
    
    log_info "Database reset completed"
}

# Check database status
check_status() {
    log_info "Checking database status..."
    
    # Test connection
    if ! test_connection; then
        return 1
    fi
    
    # Get database info
    local db_info=$(psql "$DB_CONNECTION_STRING" -t -c "
        SELECT 
            current_database() as database,
            current_user as user,
            version() as version,
            pg_size_pretty(pg_database_size(current_database())) as size;
    ")
    
    echo "Database Information:"
    echo "$db_info"
    
    # Get table count
    local table_count=$(psql "$DB_CONNECTION_STRING" -t -c "
        SELECT COUNT(*) FROM information_schema.tables 
        WHERE table_schema = 'public';
    ")
    
    echo "Tables in database: $table_count"
    
    # Get connection count
    local connection_count=$(psql "$DB_CONNECTION_STRING" -t -c "
        SELECT COUNT(*) FROM pg_stat_activity 
        WHERE datname = current_database();
    ")
    
    echo "Active connections: $connection_count"
}

# Print usage information
usage() {
    echo "Database Management Script"
    echo "Usage: $0 [command] [options]"
    echo
    echo "Commands:"
    echo "  init                 Initialize database (create and run migrations)"
    echo "  migrate              Run database migrations"
    echo "  backup [dir]         Create database backup (optional backup directory)"
    echo "  restore <file>       Restore database from backup file"
    echo "  reset                Reset database (drop and recreate)"
    echo "  status               Show database status"
    echo "  test                 Test database connectivity"
    echo "  wait                 Wait for database to be ready"
    echo
    echo "Environment Variables:"
    echo "  DB_DEPLOYMENT_MODE   Database deployment mode (container|managed|external)"
    echo "  DB_HOST              Database host"
    echo "  DB_PORT              Database port"
    echo "  DB_USER              Database user"
    echo "  DB_PASSWORD          Database password"
    echo "  DB_NAME              Database name"
    echo "  DB_SSLMODE           SSL mode (disable|require|verify-ca|verify-full)"
    echo
    echo "Examples:"
    echo "  $0 init                              # Initialize database"
    echo "  $0 backup /tmp/backups               # Create backup in /tmp/backups"
    echo "  $0 restore backup_20241101_143000.sql.gz  # Restore from backup"
    echo "  DB_DEPLOYMENT_MODE=managed $0 status # Check managed database status"
}

# Main script logic
main() {
    local command="${1:-}"
    
    case $command in
        "init")
            load_env
            get_db_connection
            wait_for_database
            create_database
            run_migrations
            ;;
        "migrate")
            load_env
            get_db_connection
            test_connection
            run_migrations
            ;;
        "backup")
            load_env
            get_db_connection
            test_connection
            create_backup "$2"
            ;;
        "restore")
            if [ -z "$2" ]; then
                log_error "Backup file not specified"
                usage
                exit 1
            fi
            load_env
            get_db_connection
            test_connection
            restore_backup "$2"
            ;;
        "reset")
            load_env
            get_db_connection
            reset_database
            ;;
        "status")
            load_env
            get_db_connection
            check_status
            ;;
        "test")
            load_env
            get_db_connection
            test_connection
            ;;
        "wait")
            load_env
            get_db_connection
            wait_for_database
            ;;
        "help"|"-h"|"--help")
            usage
            ;;
        "")
            log_error "No command specified"
            usage
            exit 1
            ;;
        *)
            log_error "Unknown command: $command"
            usage
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
