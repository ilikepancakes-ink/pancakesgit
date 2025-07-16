#!/bin/bash

# PancakesGit Setup Script
# Privacy-focused Git service with Cloudflare tunnel integration

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PANCAKESGIT_DIR="$HOME/pancakesgit"
TUNNEL_NAME="pancakesgit"
CONFIG_FILE="pancakes.yml"

print_banner() {
    echo -e "${BLUE}"
    echo "╔═══════════════════════════════════════╗"
    echo "║           PancakesGit Setup           ║"
    echo "║     Privacy-Focused Git Service      ║"
    echo "╚═══════════════════════════════════════╝"
    echo -e "${NC}"
}

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

check_dependencies() {
    log_info "Checking dependencies..."
    
    # Check if running as root
    if [[ $EUID -eq 0 ]]; then
        log_error "This script should not be run as root for security reasons"
        exit 1
    fi
    
    # Check for required tools
    local deps=("curl" "git")
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            log_error "$dep is required but not installed"
            exit 1
        fi
    done
    
    log_success "Dependencies check passed"
}

select_deployment_method() {
    echo
    log_info "Select deployment method:"
    echo "1) Docker (Recommended - Isolated and secure)"
    echo "2) Host machine (Direct installation)"
    echo
    read -p "Enter your choice (1-2): " deployment_choice
    
    case $deployment_choice in
        1)
            DEPLOYMENT_METHOD="docker"
            log_info "Selected: Docker deployment"
            ;;
        2)
            DEPLOYMENT_METHOD="host"
            log_info "Selected: Host machine deployment"
            ;;
        *)
            log_error "Invalid choice. Please run the script again."
            exit 1
            ;;
    esac
}

check_docker() {
    if [[ "$DEPLOYMENT_METHOD" == "docker" ]]; then
        if ! command -v docker &> /dev/null; then
            log_error "Docker is required but not installed"
            log_info "Please install Docker first: https://docs.docker.com/get-docker/"
            exit 1
        fi
        
        if ! command -v docker-compose &> /dev/null; then
            log_error "Docker Compose is required but not installed"
            log_info "Please install Docker Compose first"
            exit 1
        fi
        
        # Check if Docker daemon is running
        if ! docker info &> /dev/null; then
            log_error "Docker daemon is not running"
            log_info "Please start Docker first"
            exit 1
        fi
        
        log_success "Docker environment is ready"
    fi
}

setup_pancakesgit() {
    log_info "Setting up PancakesGit..."
    
    # Create directory structure
    mkdir -p "$PANCAKESGIT_DIR"
    cd "$PANCAKESGIT_DIR"
    
    if [[ "$DEPLOYMENT_METHOD" == "docker" ]]; then
        setup_docker_deployment
    else
        setup_host_deployment
    fi
    
    log_success "PancakesGit setup completed"
}

setup_docker_deployment() {
    log_info "Setting up Docker deployment..."
    
    # Create docker-compose.yml
    cat > docker-compose.yml << 'EOF'
version: '3.8'

services:
  pancakesgit:
    build: .
    container_name: pancakesgit
    restart: unless-stopped
    ports:
      - "3000:3000"
      - "2222:22"
    volumes:
      - ./data:/data
      - ./config:/config
      - ./admin_setup.env:/config/admin_setup.env:ro
      - /etc/timezone:/etc/timezone:ro
      - /etc/localtime:/etc/localtime:ro
    environment:
      - PANCAKESGIT_DOMAIN=${PANCAKESGIT_DOMAIN:-localhost}
      - PANCAKESGIT_ROOT_URL=${PANCAKESGIT_ROOT_URL:-http://localhost:3000}
      - PANCAKESGIT_SECRET_KEY=${PANCAKESGIT_SECRET_KEY}
      - PANCAKESGIT_ENCRYPTION_KEY=${PANCAKESGIT_ENCRYPTION_KEY}
      - PANCAKESGIT_DB_TYPE=postgres
      - PANCAKESGIT_DB_HOST=database
      - PANCAKESGIT_DB_PASSWORD=${DB_PASSWORD}
      - PANCAKESGIT_ENABLE_SSH=true
    depends_on:
      - database
    networks:
      - pancakesgit-network

  database:
    image: postgres:15-alpine
    container_name: pancakesgit-db
    restart: unless-stopped
    environment:
      - POSTGRES_DB=pancakesgit
      - POSTGRES_USER=pancakesgit
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - ./postgres-data:/var/lib/postgresql/data
    networks:
      - pancakesgit-network

networks:
  pancakesgit-network:
    driver: bridge
EOF
    
    # Generate secure environment variables
    generate_env_file
    
    log_success "Docker configuration created"
}

setup_host_deployment() {
    log_info "Setting up host deployment..."
    
    # Create systemd service file
    sudo tee /etc/systemd/system/pancakesgit.service > /dev/null << EOF
[Unit]
Description=PancakesGit - Privacy-focused Git Service
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$PANCAKESGIT_DIR
ExecStart=$PANCAKESGIT_DIR/pancakesgit
Restart=always
RestartSec=5
Environment=PANCAKESGIT_CONFIG=$PANCAKESGIT_DIR/config/app.ini

[Install]
WantedBy=multi-user.target
EOF
    
    sudo systemctl daemon-reload
    log_success "Systemd service created"
}

generate_env_file() {
    log_info "Generating secure environment variables..."
    
    # Generate random secrets
    SECRET_KEY=$(openssl rand -hex 32)
    ENCRYPTION_KEY=$(openssl rand -hex 32)
    DB_PASSWORD=$(openssl rand -base64 32)
    
    cat > .env << EOF
# PancakesGit Configuration
PANCAKESGIT_SECRET_KEY=$SECRET_KEY
PANCAKESGIT_ENCRYPTION_KEY=$ENCRYPTION_KEY
DB_PASSWORD=$DB_PASSWORD

# Will be set during Cloudflare tunnel setup
PANCAKESGIT_DOMAIN=
PANCAKESGIT_ROOT_URL=
EOF
    
    chmod 600 .env
    log_success "Environment file created with secure secrets"
}

check_cloudflared_login() {
    echo
    log_info "Cloudflare Tunnel Setup"
    echo "Are you already logged in to Cloudflare?"
    echo "1) Yes, I'm logged in"
    echo "2) No, I need to login"
    echo "3) Skip Cloudflare tunnel setup"
    echo
    read -p "Enter your choice (1-3): " cf_choice
    
    case $cf_choice in
        1)
            CF_LOGGED_IN=true
            ;;
        2)
            CF_LOGGED_IN=false
            ;;
        3)
            log_info "Skipping Cloudflare tunnel setup"
            return 0
            ;;
        *)
            log_error "Invalid choice"
            exit 1
            ;;
    esac
    
    setup_cloudflare_tunnel
}

setup_cloudflare_tunnel() {
    log_info "Setting up Cloudflare tunnel..."
    
    # Check if cloudflared is installed
    if ! command -v cloudflared &> /dev/null; then
        log_info "Installing cloudflared..."
        if [[ "$OSTYPE" == "linux-gnu"* ]]; then
            curl -L --output cloudflared.deb https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
            sudo dpkg -i cloudflared.deb
            rm cloudflared.deb
        elif [[ "$OSTYPE" == "darwin"* ]]; then
            brew install cloudflared
        else
            log_error "Unsupported OS. Please install cloudflared manually."
            exit 1
        fi
    fi
    
    # Login if needed
    if [[ "$CF_LOGGED_IN" == false ]]; then
        log_info "Please login to Cloudflare..."
        cloudflared tunnel login
    fi
    
    # Get domain from user
    echo
    read -p "Enter the domain you want to use for PancakesGit: " DOMAIN
    
    if [[ -z "$DOMAIN" ]]; then
        log_error "Domain cannot be empty"
        exit 1
    fi
    
    # Create tunnel
    log_info "Creating tunnel '$TUNNEL_NAME'..."
    cloudflared tunnel create "$TUNNEL_NAME"
    
    # Get tunnel ID
    TUNNEL_ID=$(cloudflared tunnel list | grep "$TUNNEL_NAME" | awk '{print $1}')
    
    if [[ -z "$TUNNEL_ID" ]]; then
        log_error "Failed to create tunnel"
        exit 1
    fi
    
    # Create custom config file
    mkdir -p ~/.cloudflared
    cat > ~/.cloudflared/$CONFIG_FILE << EOF
tunnel: $TUNNEL_ID
credentials-file: ~/.cloudflared/$TUNNEL_ID.json

ingress:
  - hostname: $DOMAIN
    service: http://localhost:3000
  - service: http_status:404
EOF
    
    # Create DNS record
    log_info "Creating DNS record..."
    cloudflared tunnel route dns "$TUNNEL_NAME" "$DOMAIN"
    
    # Update environment file with domain
    if [[ -f .env ]]; then
        sed -i "s/PANCAKESGIT_DOMAIN=/PANCAKESGIT_DOMAIN=$DOMAIN/" .env
        sed -i "s|PANCAKESGIT_ROOT_URL=|PANCAKESGIT_ROOT_URL=https://$DOMAIN|" .env
    fi
    
    log_success "Cloudflare tunnel '$TUNNEL_NAME' created successfully!"
    echo
    log_info "To start the tunnel, run:"
    echo -e "${GREEN}cloudflared tunnel --config ~/.cloudflared/$CONFIG_FILE run${NC}"
    echo
    log_info "Your PancakesGit instance will be available at: https://$DOMAIN"
}

create_admin_account() {
    echo
    log_info "Creating admin account..."
    echo "Please create an admin account for PancakesGit:"
    echo

    # Get admin details
    while true; do
        read -p "Admin username: " ADMIN_USERNAME
        if [[ -n "$ADMIN_USERNAME" && "$ADMIN_USERNAME" =~ ^[a-zA-Z0-9_-]+$ ]]; then
            break
        else
            log_error "Username must contain only letters, numbers, underscores, and hyphens"
        fi
    done

    while true; do
        read -p "Admin email: " ADMIN_EMAIL
        if [[ "$ADMIN_EMAIL" =~ ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
            break
        else
            log_error "Please enter a valid email address"
        fi
    done

    while true; do
        read -s -p "Admin password: " ADMIN_PASSWORD
        echo
        if [[ ${#ADMIN_PASSWORD} -ge 8 ]]; then
            read -s -p "Confirm password: " ADMIN_PASSWORD_CONFIRM
            echo
            if [[ "$ADMIN_PASSWORD" == "$ADMIN_PASSWORD_CONFIRM" ]]; then
                break
            else
                log_error "Passwords do not match"
            fi
        else
            log_error "Password must be at least 8 characters long"
        fi
    done

    # Save admin credentials to a temporary file for container initialization
    cat > admin_setup.env << EOF
ADMIN_USERNAME=$ADMIN_USERNAME
ADMIN_EMAIL=$ADMIN_EMAIL
ADMIN_PASSWORD=$ADMIN_PASSWORD
EOF

    chmod 600 admin_setup.env
    log_success "Admin account configuration saved"
}

show_completion_message() {
    echo
    log_success "PancakesGit setup completed!"
    echo
    echo -e "${BLUE}Next steps:${NC}"

    if [[ "$DEPLOYMENT_METHOD" == "docker" ]]; then
        echo "1. Start PancakesGit:"
        echo -e "   ${GREEN}cd $PANCAKESGIT_DIR && docker-compose up -d${NC}"
        echo
        echo "2. Check logs to ensure everything started correctly:"
        echo -e "   ${GREEN}docker-compose logs -f pancakesgit${NC}"
        echo
        echo "   The admin account will be created automatically during startup."
    else
        echo "1. Build PancakesGit:"
        echo -e "   ${GREEN}cd $PANCAKESGIT_DIR && go build -o pancakesgit ./cmd/pancakesgit${NC}"
        echo
        echo "2. Run database migrations:"
        echo -e "   ${GREEN}./pancakesgit migrate${NC}"
        echo
        echo "3. Create admin account:"
        echo -e "   ${GREEN}ADMIN_USERNAME=$ADMIN_USERNAME ADMIN_EMAIL=$ADMIN_EMAIL ADMIN_PASSWORD=$ADMIN_PASSWORD ADMIN_CREATE=true ./pancakesgit create-user${NC}"
        echo
        echo "4. Start PancakesGit:"
        echo -e "   ${GREEN}sudo systemctl enable pancakesgit && sudo systemctl start pancakesgit${NC}"
    fi

    if [[ -f ~/.cloudflared/$CONFIG_FILE ]]; then
        echo
        echo "3. Start Cloudflare tunnel:"
        echo -e "   ${GREEN}cloudflared tunnel --config ~/.cloudflared/$CONFIG_FILE run${NC}"
        echo
        echo "4. Access your instance at: https://$DOMAIN"
    else
        echo
        echo "3. Access your instance at: http://localhost:3000"
    fi

    echo
    echo -e "${YELLOW}Admin Account Details:${NC}"
    echo "- Username: $ADMIN_USERNAME"
    echo "- Email: $ADMIN_EMAIL"
    echo "- Login at your instance URL to get started"
    echo
    echo -e "${YELLOW}Important:${NC}"
    echo "- Your encryption keys are stored in $PANCAKESGIT_DIR/.env"
    echo "- Keep this file secure and backed up"
    echo "- Admin credentials are in $PANCAKESGIT_DIR/admin_setup.env (delete after first login)"
}

main() {
    print_banner
    check_dependencies
    select_deployment_method
    check_docker
    setup_pancakesgit
    create_admin_account
    check_cloudflared_login
    show_completion_message
}

# Run main function
main "$@"
