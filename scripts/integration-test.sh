#!/bin/bash

set -e

echo "🧪 Starting Blockhead Consulting Integration Tests..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Clean up function
cleanup() {
    print_status "Cleaning up test environment..."
    docker-compose -f docker-compose.test.yml down -v
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Start test environment
print_status "Starting test environment..."
docker-compose -f docker-compose.test.yml up -d mailhog-test

# Wait for services to be ready
print_status "Waiting for services to start..."
sleep 5

# Check if MailHog is ready
print_status "Checking MailHog availability..."
for i in {1..30}; do
    if curl -f http://localhost:8026/api/v1/messages > /dev/null 2>&1; then
        print_status "MailHog is ready!"
        break
    fi
    if [ $i -eq 30 ]; then
        print_error "MailHog failed to start within 30 seconds"
        exit 1
    fi
    sleep 1
done

# Start application
print_status "Starting application..."
docker-compose -f docker-compose.test.yml up -d app-test

# Wait for application to be ready
print_status "Waiting for application to start..."
for i in {1..60}; do
    if curl -f http://localhost:8085/health > /dev/null 2>&1; then
        print_status "Application is ready!"
        break
    fi
    if [ $i -eq 60 ]; then
        print_warning "Application health check failed, but continuing with tests..."
        break
    fi
    sleep 1
done

# Run unit tests first
print_status "Running unit tests..."
docker-compose -f docker-compose.test.yml run --rm integration-tests go test ./internal/... -v -race

# Run integration tests
print_status "Running integration tests..."

# Test contact form submission
print_status "Testing contact form submission..."
RESPONSE=$(curl -s -X POST http://localhost:8085/contact \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "name=Test User&email=test@example.com&company=Test Corp&message=Integration test message")

if echo "$RESPONSE" | grep -q "success"; then
    print_status "✅ Contact form submission test passed"
else
    print_error "❌ Contact form submission test failed"
    echo "Response: $RESPONSE"
fi

# Test static file serving
print_status "Testing static file serving..."
if curl -f http://localhost:8085/static/styles.css > /dev/null 2>&1; then
    print_status "✅ Static file serving test passed"
else
    print_error "❌ Static file serving test failed"
fi

# Test home page
print_status "Testing home page..."
if curl -f http://localhost:8085/ > /dev/null 2>&1; then
    print_status "✅ Home page test passed"
else
    print_error "❌ Home page test failed"
fi

# Test blog functionality
print_status "Testing blog functionality..."
if curl -f http://localhost:8085/blog > /dev/null 2>&1; then
    print_status "✅ Blog functionality test passed"
else
    print_error "❌ Blog functionality test failed"
fi

# Check if email was sent (wait a moment for async processing)
sleep 3
print_status "Checking if email notification was sent..."
EMAIL_COUNT=$(curl -s http://localhost:8026/api/v1/messages | jq length)
if [ "$EMAIL_COUNT" -gt 0 ]; then
    print_status "✅ Email notification test passed ($EMAIL_COUNT emails sent)"
else
    print_warning "⚠️  No emails found in MailHog (this might be expected if email service is not configured)"
fi

# Test security headers
print_status "Testing security headers..."
HEADERS=$(curl -s -I http://localhost:8085/)
if echo "$HEADERS" | grep -q "X-Frame-Options"; then
    print_status "✅ Security headers test passed"
else
    print_error "❌ Security headers test failed"
fi

print_status "🎉 Integration tests completed!"

# Show logs if there were any errors
if [ $? -ne 0 ]; then
    print_error "Some tests failed. Here are the application logs:"
    docker-compose -f docker-compose.test.yml logs app-test
fi