#!/bin/bash

# Test script for signaling server
# Usage: ./scripts/test.sh

set -e

echo "=== Signaling Server Test Suite ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}➜ $1${NC}"
}

# Check if server is running
check_server() {
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Test 1: Build
print_info "Test 1: Building server..."
if go build -o /tmp/signaling-server ./cmd/server; then
    print_success "Build successful"
else
    print_error "Build failed"
    exit 1
fi

# Test 2: Start server
print_info "Test 2: Starting server..."
/tmp/signaling-server &
SERVER_PID=$!
sleep 2

if check_server; then
    print_success "Server started (PID: $SERVER_PID)"
else
    print_error "Server failed to start"
    kill $SERVER_PID 2>/dev/null || true
    exit 1
fi

# Test 3: Health check
print_info "Test 3: Testing health endpoint..."
HEALTH_RESPONSE=$(curl -s http://localhost:8080/health)
if [[ $HEALTH_RESPONSE == *"healthy"* ]]; then
    print_success "Health check passed"
else
    print_error "Health check failed"
    kill $SERVER_PID
    exit 1
fi

# Test 4: WebSocket connection
print_info "Test 4: Testing WebSocket connection..."
timeout 5 go run ./cmd/testclient/host.go -session=TEST_$(date +%s) > /tmp/host_test.log 2>&1 &
HOST_PID=$!
sleep 1

if ps -p $HOST_PID > /dev/null; then
    print_success "WebSocket connection successful"
    kill $HOST_PID 2>/dev/null || true
else
    print_error "WebSocket connection failed"
    cat /tmp/host_test.log
    kill $SERVER_PID
    exit 1
fi

# Test 5: Unit tests
print_info "Test 5: Running unit tests..."
if go test ./... -cover; then
    print_success "Unit tests passed"
else
    print_error "Unit tests failed"
    kill $SERVER_PID
    exit 1
fi

# Cleanup
print_info "Cleaning up..."
kill $SERVER_PID 2>/dev/null || true
rm -f /tmp/signaling-server /tmp/host_test.log

echo ""
echo -e "${GREEN}=== All tests passed! ===${NC}"
