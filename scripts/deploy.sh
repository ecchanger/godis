#!/bin/bash

# Deployment script for Godis Web Todo Server
# Supports Docker Compose deployment with different profiles

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

show_usage() {
    cat << EOF
Usage: $0 [COMMAND] [OPTIONS]

Deploy and manage Godis Web Todo Server

COMMANDS:
    start       Start the services
    stop        Stop the services
    restart     Restart the services
    status      Show service status
    logs        Show service logs
    update      Update and restart services
    clean       Clean up containers and volumes

OPTIONS:
    --profile PROFILE   Docker Compose profile (default: none)
                       Available: production, monitoring
    --build            Force rebuild images
    --detach           Run in background (default for start)
    --follow           Follow logs output
    --help             Show this help

EXAMPLES:
    $0 start                    # Start basic services
    $0 start --profile production    # Start with Nginx proxy
    $0 start --profile monitoring   # Start with Redis monitoring
    $0 logs --follow            # Follow logs
    $0 update --build           # Update with rebuild

EOF
}

# Default values
COMMAND=""
PROFILE=""
BUILD_FLAG=""
DETACH_FLAG="-d"
FOLLOW_FLAG=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        start|stop|restart|status|logs|update|clean)
            COMMAND="$1"
            shift
            ;;
        --profile)
            PROFILE="$2"
            shift 2
            ;;
        --build)
            BUILD_FLAG="--build"
            shift
            ;;
        --detach)
            DETACH_FLAG="-d"
            shift
            ;;
        --follow)
            FOLLOW_FLAG="-f"
            shift
            ;;
        --help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Ensure directories exist
ensure_directories() {
    print_status "Ensuring required directories exist..."
    mkdir -p data logs ssl
    
    # Set appropriate permissions
    chmod 755 data logs
    
    if [ -d ssl ]; then
        chmod 700 ssl
    fi
}

# Generate self-signed certificates for development
generate_ssl_certs() {
    if [ ! -f ssl/cert.pem ] || [ ! -f ssl/key.pem ]; then
        print_status "Generating self-signed SSL certificates..."
        mkdir -p ssl
        
        openssl req -x509 -newkey rsa:4096 -keyout ssl/key.pem -out ssl/cert.pem \
            -days 365 -nodes -subj "/CN=localhost"
        
        chmod 600 ssl/key.pem ssl/cert.pem
        print_success "SSL certificates generated"
    fi
}

# Build Docker Compose command
build_compose_cmd() {
    local cmd="docker-compose"
    
    if [ -n "$PROFILE" ]; then
        cmd="$cmd --profile $PROFILE"
    fi
    
    echo "$cmd"
}

# Start services
start_services() {
    print_status "Starting Godis Web Todo services..."
    ensure_directories
    
    if [ "$PROFILE" = "production" ]; then
        generate_ssl_certs
    fi
    
    local compose_cmd=$(build_compose_cmd)
    $compose_cmd up $BUILD_FLAG $DETACH_FLAG
    
    if [ "$?" -eq 0 ]; then
        print_success "Services started successfully"
        print_status "Web Todo interface: http://localhost:8080"
        print_status "Redis server: localhost:6399"
        
        if [ "$PROFILE" = "production" ]; then
            print_status "Nginx proxy: http://localhost:80, https://localhost:443"
        fi
        
        if [ "$PROFILE" = "monitoring" ]; then
            print_status "Redis Insight: http://localhost:8001"
        fi
    else
        print_error "Failed to start services"
        exit 1
    fi
}

# Stop services
stop_services() {
    print_status "Stopping Godis Web Todo services..."
    local compose_cmd=$(build_compose_cmd)
    $compose_cmd down
    
    if [ "$?" -eq 0 ]; then
        print_success "Services stopped successfully"
    else
        print_error "Failed to stop services"
        exit 1
    fi
}

# Restart services
restart_services() {
    print_status "Restarting Godis Web Todo services..."
    stop_services
    start_services
}

# Show service status
show_status() {
    print_status "Service status:"
    local compose_cmd=$(build_compose_cmd)
    $compose_cmd ps
}

# Show logs
show_logs() {
    print_status "Showing service logs..."
    local compose_cmd=$(build_compose_cmd)
    $compose_cmd logs $FOLLOW_FLAG
}

# Update services
update_services() {
    print_status "Updating Godis Web Todo services..."
    local compose_cmd=$(build_compose_cmd)
    
    # Pull latest images
    $compose_cmd pull
    
    # Restart with build if requested
    $compose_cmd up $BUILD_FLAG $DETACH_FLAG --force-recreate
    
    if [ "$?" -eq 0 ]; then
        print_success "Services updated successfully"
    else
        print_error "Failed to update services"
        exit 1
    fi
}

# Clean up
clean_services() {
    print_warning "This will remove all containers, networks, and volumes!"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_status "Cleaning up services..."
        local compose_cmd=$(build_compose_cmd)
        $compose_cmd down -v --remove-orphans
        
        # Remove unused images
        docker system prune -f
        
        print_success "Cleanup completed"
    else
        print_status "Cleanup cancelled"
    fi
}

# Main execution
case "$COMMAND" in
    start)
        start_services
        ;;
    stop)
        stop_services
        ;;
    restart)
        restart_services
        ;;
    status)
        show_status
        ;;
    logs)
        show_logs
        ;;
    update)
        update_services
        ;;
    clean)
        clean_services
        ;;
    "")
        print_error "No command specified"
        show_usage
        exit 1
        ;;
    *)
        print_error "Unknown command: $COMMAND"
        show_usage
        exit 1
        ;;
esac