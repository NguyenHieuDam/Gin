#!/bin/bash

# News Aggregator Deployment Script
# Usage: ./deploy.sh [production|development]

set -e

ENVIRONMENT=${1:-development}
PROJECT_NAME="news-aggregator"

echo "🚀 Deploying News Aggregator - Environment: $ENVIRONMENT"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
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

# Check if Docker is installed
check_docker() {
    print_status "Checking Docker installation..."
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    print_success "Docker and Docker Compose are installed"
}

# Check if required files exist
check_files() {
    print_status "Checking required files..."
    
    local required_files=("go.mod" "docker-compose.yml")
    if [ "$ENVIRONMENT" = "production" ]; then
        required_files+=("docker-compose.production.yml" "Dockerfile.production")
    fi
    
    for file in "${required_files[@]}"; do
        if [ ! -f "$file" ]; then
            print_error "Required file $file not found"
            exit 1
        fi
    done
    
    print_success "All required files found"
}

# Build and start services
deploy_services() {
    print_status "Building and starting services..."
    
    if [ "$ENVIRONMENT" = "production" ]; then
        print_status "Deploying in PRODUCTION mode"
        docker-compose -f docker-compose.production.yml down --remove-orphans
        docker-compose -f docker-compose.production.yml build --no-cache
        docker-compose -f docker-compose.production.yml up -d
    else
        print_status "Deploying in DEVELOPMENT mode"
        docker-compose down --remove-orphans
        docker-compose build --no-cache
        docker-compose up -d
    fi
    
    print_success "Services started successfully"
}

# Wait for services to be ready
wait_for_services() {
    print_status "Waiting for services to be ready..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            print_success "News API is ready"
            break
        fi
        
        if [ $attempt -eq $max_attempts ]; then
            print_error "News API failed to start within expected time"
            exit 1
        fi
        
        print_status "Waiting for News API... (attempt $attempt/$max_attempts)"
        sleep 5
        ((attempt++))
    done
    
    # Check collector
    attempt=1
    while [ $attempt -le $max_attempts ]; do
        if curl -s http://localhost:8081/health > /dev/null 2>&1; then
            print_success "News Collector is ready"
            break
        fi
        
        if [ $attempt -eq $max_attempts ]; then
            print_warning "News Collector may not be ready yet"
            break
        fi
        
        print_status "Waiting for News Collector... (attempt $attempt/$max_attempts)"
        sleep 5
        ((attempt++))
    done
}

# Show deployment status
show_status() {
    print_status "Deployment Status:"
    echo ""
    
    if [ "$ENVIRONMENT" = "production" ]; then
        docker-compose -f docker-compose.production.yml ps
    else
        docker-compose ps
    fi
    
    echo ""
    print_status "Service URLs:"
    echo "  🌐 Web Interface: http://localhost"
    echo "  📡 News API: http://localhost:8080"
    echo "  🔄 News Collector: http://localhost:8081"
    echo "  📊 Kafka UI: http://localhost:8082"
    echo ""
    
    print_status "Health Checks:"
    echo "  API Health: $(curl -s http://localhost:8080/health | jq -r '.status' 2>/dev/null || echo 'Unknown')"
    echo "  Collector Health: $(curl -s http://localhost:8081/health | jq -r '.status' 2>/dev/null || echo 'Unknown')"
}

# Collect some initial news
collect_initial_news() {
    print_status "Collecting initial news..."
    
    if curl -s -X POST http://localhost:8081/collect > /dev/null 2>&1; then
        print_success "Initial news collection completed"
    else
        print_warning "Failed to collect initial news (this is normal if collector is still starting)"
    fi
}

# Show logs
show_logs() {
    print_status "Recent logs:"
    echo ""
    
    if [ "$ENVIRONMENT" = "production" ]; then
        docker-compose -f docker-compose.production.yml logs --tail=20
    else
        docker-compose logs --tail=20
    fi
}

# Cleanup function
cleanup() {
    print_status "Cleaning up..."
    if [ "$ENVIRONMENT" = "production" ]; then
        docker-compose -f docker-compose.production.yml down
    else
        docker-compose down
    fi
    print_success "Cleanup completed"
}

# Main deployment function
main() {
    print_status "Starting News Aggregator deployment..."
    echo ""
    
    # Pre-deployment checks
    check_docker
    check_files
    
    # Deploy services
    deploy_services
    
    # Wait for services
    wait_for_services
    
    # Collect initial news
    collect_initial_news
    
    # Show status
    show_status
    
    echo ""
    print_success "🎉 Deployment completed successfully!"
    echo ""
    print_status "Next steps:"
    echo "  1. Open http://localhost in your browser"
    echo "  2. Check the web interface"
    echo "  3. Monitor logs with: docker-compose logs -f"
    echo "  4. To stop: docker-compose down"
    echo ""
    
    # Ask if user wants to see logs
    read -p "Do you want to see recent logs? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        show_logs
    fi
}

# Handle script interruption
trap cleanup INT TERM

# Run main function
main "$@"
