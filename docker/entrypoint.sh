#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Wait for database to be ready
wait_for_db() {
    if [[ "$PANCAKESGIT_DB_TYPE" == "postgres" ]]; then
        log_info "Waiting for PostgreSQL database..."
        while ! pg_isready -h database -p 5432 -U pancakesgit; do
            sleep 2
        done
        log_success "PostgreSQL is ready"
    fi
}

# Initialize database
init_database() {
    log_info "Initializing database..."
    
    # Run migrations
    ./pancakesgit migrate
    
    log_success "Database initialized"
}

# Create admin user if credentials are provided
create_admin_user() {
    if [[ -f /config/admin_setup.env ]]; then
        log_info "Creating admin user..."
        
        # Source admin credentials
        source /config/admin_setup.env
        
        # Create admin user using environment variables
        export ADMIN_CREATE=true
        ./pancakesgit create-user
        
        # Remove admin setup file for security
        rm -f /config/admin_setup.env
        
        log_success "Admin user created successfully"
    else
        log_warning "No admin setup file found. Admin user must be created manually."
    fi
}

# Setup SSH host keys
setup_ssh() {
    if [[ "$PANCAKESGIT_ENABLE_SSH" == "true" ]]; then
        log_info "Setting up SSH..."
        
        # Generate SSH host keys if they don't exist
        if [[ ! -f /data/ssh/ssh_host_rsa_key ]]; then
            mkdir -p /data/ssh
            ssh-keygen -t rsa -b 4096 -f /data/ssh/ssh_host_rsa_key -N ""
            ssh-keygen -t ed25519 -f /data/ssh/ssh_host_ed25519_key -N ""
            chmod 600 /data/ssh/ssh_host_*_key
            chmod 644 /data/ssh/ssh_host_*_key.pub
        fi
        
        log_success "SSH setup completed"
    fi
}

# Setup directories and permissions
setup_directories() {
    log_info "Setting up directories..."
    
    # Create necessary directories
    mkdir -p /data/repositories
    mkdir -p /data/uploads
    mkdir -p /data/logs
    mkdir -p /data/tmp
    
    # Set permissions
    chmod 755 /data/repositories
    chmod 755 /data/uploads
    chmod 755 /data/logs
    chmod 755 /data/tmp
    
    log_success "Directories setup completed"
}

# Generate default configuration if it doesn't exist
setup_config() {
    if [[ ! -f /config/app.ini ]]; then
        log_info "Generating default configuration..."
        
        cat > /config/app.ini << EOF
[server]
host = 0.0.0.0
port = 3000
domain = ${PANCAKESGIT_DOMAIN:-localhost}
root_url = ${PANCAKESGIT_ROOT_URL:-http://localhost:3000}
debug = false

[database]
type = ${PANCAKESGIT_DB_TYPE:-postgres}
host = ${PANCAKESGIT_DB_HOST:-database}
port = ${PANCAKESGIT_DB_PORT:-5432}
name = ${PANCAKESGIT_DB_NAME:-pancakesgit}
user = ${PANCAKESGIT_DB_USER:-pancakesgit}
password = ${PANCAKESGIT_DB_PASSWORD}
ssl_mode = ${PANCAKESGIT_DB_SSL_MODE:-disable}

[security]
secret_key = ${PANCAKESGIT_SECRET_KEY}
session_timeout = 3600
password_min_length = 8
require_strong_password = true
enable_2fa = true
csrf_protection = true
secure_cookies = true

[git]
enable_http = true
enable_ssh = ${PANCAKESGIT_ENABLE_SSH:-true}
ssh_port = 2222
max_repo_size = 1073741824

[privacy]
anonymous_access = false
hide_user_emails = true
disable_gravatar = true
log_retention_days = 30
enable_tor_access = true
disable_analytics = true
minimize_metadata = true

[storage]
type = local
path = /data/repositories
encrypt_at_rest = true

[encryption]
key = ${PANCAKESGIT_ENCRYPTION_KEY}
algorithm = AES-256-GCM
encrypt_repositories = true
encrypt_database = true
encrypt_logs = true
EOF
        
        log_success "Configuration generated"
    fi
}

# Health check function
health_check() {
    # Simple health check - verify the binary exists and config is readable
    if [[ -x "./pancakesgit" && -r "/config/app.ini" ]]; then
        exit 0
    else
        exit 1
    fi
}

# Main initialization
main() {
    log_info "Starting PancakesGit initialization..."
    
    # Handle health check
    if [[ "$1" == "health" ]]; then
        health_check
        exit $?
    fi
    
    # Setup
    setup_directories
    setup_config
    wait_for_db
    init_database
    create_admin_user
    setup_ssh
    
    log_success "PancakesGit initialization completed"
    
    # Start the application
    log_info "Starting PancakesGit server..."
    exec "$@"
}

# Run main function with all arguments
main "$@"
