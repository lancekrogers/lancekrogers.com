#!/bin/bash

set -e

echo "🚀 Setting up Blockhead Consulting Development Environment..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_info() {
    echo -e "${BLUE}[DEV]${NC} $1"
}

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    print_error "Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# Create necessary directories
print_status "Creating development directories..."
mkdir -p data/messages
mkdir -p test_results

# Create development .env file if it doesn't exist
if [ ! -f .env.development ]; then
    print_status "Creating development environment file..."
    cat > .env.development << EOF
# Development environment configuration
ENVIRONMENT=development
PORT=8085
CALENDAR_ENABLED=false
BLOG_ENABLED=true
HERO_STYLE=professional

# Git storage configuration
GIT_REPO_PATH=./data/messages
GIT_ENCRYPTION_KEY=dev_key_0123456789abcdef0123456789abcdef
GIT_BRANCH=main
GIT_COMMIT_AUTHOR=Dev Bot
GIT_COMMIT_EMAIL=dev@blockhead.consulting
GIT_PUSH_ON_WRITE=false

# Email configuration (MailHog for development)
SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_USERNAME=dev@blockhead.consulting
SMTP_PASSWORD=devpass
SMTP_FROM_ADDRESS=noreply@blockhead.consulting
SMTP_FROM_NAME=Blockhead Consulting Dev
SMTP_TLS_ENABLED=false
EOF
    print_status "Created .env.development file"
else
    print_info ".env.development already exists"
fi

# Build Docker images
print_status "Building Docker images..."
docker-compose build

print_status "Starting development services..."
docker-compose up -d mailhog gitea

# Wait for services to be ready
print_status "Waiting for services to start..."
sleep 5

# Check service status
print_info "Development environment is ready!"
print_info ""
print_info "Available services:"
print_info "📧 MailHog (Email testing): http://localhost:8025"
print_info "🗂️  Gitea (Git server): http://localhost:3000"
print_info "📱 Redis: localhost:6379"
print_info ""
print_info "To start the application:"
print_info "  docker-compose up app"
print_info ""
print_info "To run integration tests:"
print_info "  ./scripts/integration-test.sh"
print_info ""
print_info "To view logs:"
print_info "  docker-compose logs -f [service-name]"
print_info ""
print_info "To stop all services:"
print_info "  docker-compose down"

print_status "🎉 Development environment setup complete!"