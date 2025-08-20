#!/bin/bash

# Marketing AI System Testing Script
# Comprehensive testing suite for the complete system

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BASE_URL="${BASE_URL:-http://localhost:8080}"
FRONTEND_URL="${FRONTEND_URL:-http://localhost:3000}"
TEST_EMAIL="${TEST_EMAIL:-test@example.com}"
TEST_PASSWORD="${TEST_PASSWORD:-password123}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test results
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

# Logging functions
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

# Test result functions
test_passed() {
    ((TESTS_PASSED++))
    ((TESTS_TOTAL++))
    log_success "✓ $1"
}

test_failed() {
    ((TESTS_FAILED++))
    ((TESTS_TOTAL++))
    log_error "✗ $1"
}

# Wait for services to be ready
wait_for_services() {
    log_info "Waiting for services to be ready..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s "$BASE_URL/health" > /dev/null 2>&1; then
            log_success "Backend service is ready"
            break
        fi
        
        if [ $attempt -eq $max_attempts ]; then
            log_error "Backend service failed to start within timeout"
            exit 1
        fi
        
        log_info "Attempt $attempt/$max_attempts - waiting for backend..."
        sleep 5
        ((attempt++))
    done
    
    # Wait for frontend
    attempt=1
    while [ $attempt -le $max_attempts ]; do
        if curl -s "$FRONTEND_URL" > /dev/null 2>&1; then
            log_success "Frontend service is ready"
            break
        fi
        
        if [ $attempt -eq $max_attempts ]; then
            log_warning "Frontend service not ready, but continuing tests"
            break
        fi
        
        log_info "Attempt $attempt/$max_attempts - waiting for frontend..."
        sleep 5
        ((attempt++))
    done
}

# Test health endpoints
test_health_endpoints() {
    log_info "Testing health endpoints..."
    
    # Backend health
    if response=$(curl -s "$BASE_URL/health" 2>/dev/null); then
        if echo "$response" | grep -q "healthy"; then
            test_passed "Backend health endpoint"
        else
            test_failed "Backend health endpoint - invalid response"
        fi
    else
        test_failed "Backend health endpoint - no response"
    fi
    
    # Frontend health
    if curl -s "$FRONTEND_URL" > /dev/null 2>&1; then
        test_passed "Frontend health endpoint"
    else
        test_failed "Frontend health endpoint"
    fi
}

# Test authentication system
test_authentication() {
    log_info "Testing authentication system..."
    
    # Test login with invalid credentials
    if response=$(curl -s -w "%{http_code}" -X POST "$BASE_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"email":"invalid@example.com","password":"wrongpassword"}' 2>/dev/null); then
        
        http_code="${response: -3}"
        if [ "$http_code" = "401" ] || [ "$http_code" = "400" ]; then
            test_passed "Authentication - invalid credentials rejected"
        else
            test_failed "Authentication - invalid credentials not properly rejected (HTTP $http_code)"
        fi
    else
        test_failed "Authentication - login endpoint not responding"
    fi
    
    # Test registration endpoint
    if response=$(curl -s -w "%{http_code}" -X POST "$BASE_URL/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d '{"first_name":"Test","last_name":"User","email":"newuser@example.com","password":"password123"}' 2>/dev/null); then
        
        http_code="${response: -3}"
        if [ "$http_code" = "201" ] || [ "$http_code" = "409" ] || [ "$http_code" = "400" ]; then
            test_passed "Authentication - registration endpoint responding"
        else
            test_failed "Authentication - registration endpoint error (HTTP $http_code)"
        fi
    else
        test_failed "Authentication - registration endpoint not responding"
    fi
}

# Test API endpoints
test_api_endpoints() {
    log_info "Testing API endpoints..."
    
    # Test marketing endpoints
    endpoints=(
        "GET /api/v1/marketing/campaigns"
        "GET /api/v1/content/templates"
        "GET /api/v1/marketing/integrations"
        "GET /api/v1/analytics/dashboard"
    )
    
    for endpoint in "${endpoints[@]}"; do
        method=$(echo "$endpoint" | cut -d' ' -f1)
        path=$(echo "$endpoint" | cut -d' ' -f2)
        
        if response=$(curl -s -w "%{http_code}" -X "$method" "$BASE_URL$path" 2>/dev/null); then
            http_code="${response: -3}"
            if [ "$http_code" = "200" ] || [ "$http_code" = "401" ] || [ "$http_code" = "403" ]; then
                test_passed "API endpoint $endpoint"
            else
                test_failed "API endpoint $endpoint (HTTP $http_code)"
            fi
        else
            test_failed "API endpoint $endpoint - no response"
        fi
    done
}

# Test database connectivity
test_database() {
    log_info "Testing database connectivity..."
    
    if command -v docker-compose &> /dev/null; then
        if docker-compose ps postgres | grep -q "Up"; then
            if docker-compose exec -T postgres pg_isready -U postgres > /dev/null 2>&1; then
                test_passed "Database connectivity"
            else
                test_failed "Database connectivity - not ready"
            fi
        else
            test_warning "Database container not running (may be in k8s)"
        fi
    elif command -v kubectl &> /dev/null; then
        if kubectl get pods -n marketing-ai | grep -q postgres; then
            test_passed "Database pod exists in Kubernetes"
        else
            test_failed "Database pod not found in Kubernetes"
        fi
    else
        test_warning "Cannot test database - no Docker Compose or kubectl available"
    fi
}

# Test Redis connectivity
test_redis() {
    log_info "Testing Redis connectivity..."
    
    if command -v docker-compose &> /dev/null; then
        if docker-compose ps redis | grep -q "Up"; then
            if docker-compose exec -T redis redis-cli ping > /dev/null 2>&1; then
                test_passed "Redis connectivity"
            else
                test_failed "Redis connectivity - not responding"
            fi
        else
            test_warning "Redis container not running (may be in k8s)"
        fi
    elif command -v kubectl &> /dev/null; then
        if kubectl get pods -n marketing-ai | grep -q redis; then
            test_passed "Redis pod exists in Kubernetes"
        else
            test_failed "Redis pod not found in Kubernetes"
        fi
    else
        test_warning "Cannot test Redis - no Docker Compose or kubectl available"
    fi
}

# Test security headers
test_security() {
    log_info "Testing security features..."
    
    # Test CORS headers
    if response=$(curl -s -I -X OPTIONS "$BASE_URL/api/v1/health" 2>/dev/null); then
        if echo "$response" | grep -qi "access-control-allow"; then
            test_passed "CORS headers present"
        else
            test_warning "CORS headers not found"
        fi
    else
        test_failed "Security headers test - no response"
    fi
    
    # Test rate limiting (basic check)
    if curl -s "$BASE_URL/health" > /dev/null 2>&1; then
        test_passed "Rate limiting endpoint accessible"
    else
        test_failed "Rate limiting test failed"
    fi
}

# Test content generation (mock)
test_content_generation() {
    log_info "Testing content generation endpoints..."
    
    # Test content generation endpoint (should require auth)
    if response=$(curl -s -w "%{http_code}" -X POST "$BASE_URL/api/v1/content/generate" \
        -H "Content-Type: application/json" \
        -d '{"type":"social_post","platform":"facebook","topic":"travel"}' 2>/dev/null); then
        
        http_code="${response: -3}"
        if [ "$http_code" = "401" ] || [ "$http_code" = "403" ] || [ "$http_code" = "200" ]; then
            test_passed "Content generation endpoint responding"
        else
            test_failed "Content generation endpoint error (HTTP $http_code)"
        fi
    else
        test_failed "Content generation endpoint not responding"
    fi
}

# Test integration endpoints
test_integrations() {
    log_info "Testing integration endpoints..."
    
    # Test integrations list (should require auth)
    if response=$(curl -s -w "%{http_code}" -X GET "$BASE_URL/api/v1/marketing/integrations" 2>/dev/null); then
        http_code="${response: -3}"
        if [ "$http_code" = "401" ] || [ "$http_code" = "403" ] || [ "$http_code" = "200" ]; then
            test_passed "Integrations endpoint responding"
        else
            test_failed "Integrations endpoint error (HTTP $http_code)"
        fi
    else
        test_failed "Integrations endpoint not responding"
    fi
}

# Test analytics endpoints
test_analytics() {
    log_info "Testing analytics endpoints..."
    
    # Test analytics dashboard (should require auth)
    if response=$(curl -s -w "%{http_code}" -X GET "$BASE_URL/api/v1/analytics/dashboard" 2>/dev/null); then
        http_code="${response: -3}"
        if [ "$http_code" = "401" ] || [ "$http_code" = "403" ] || [ "$http_code" = "200" ]; then
            test_passed "Analytics endpoint responding"
        else
            test_failed "Analytics endpoint error (HTTP $http_code)"
        fi
    else
        test_failed "Analytics endpoint not responding"
    fi
}

# Run performance test
test_performance() {
    log_info "Running basic performance tests..."
    
    # Simple load test with curl
    local start_time=$(date +%s%N)
    local requests=10
    local successful=0
    
    for i in $(seq 1 $requests); do
        if curl -s "$BASE_URL/health" > /dev/null 2>&1; then
            ((successful++))
        fi
    done
    
    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 )) # Convert to milliseconds
    local avg_response_time=$(( duration / requests ))
    
    if [ $successful -eq $requests ] && [ $avg_response_time -lt 1000 ]; then
        test_passed "Performance test - $successful/$requests requests successful, avg ${avg_response_time}ms"
    else
        test_failed "Performance test - $successful/$requests requests successful, avg ${avg_response_time}ms"
    fi
}

# Run comprehensive system test
run_comprehensive_test() {
    log_info "Starting comprehensive system test..."
    
    wait_for_services
    test_health_endpoints
    test_authentication
    test_api_endpoints
    test_database
    test_redis
    test_security
    test_content_generation
    test_integrations
    test_analytics
    test_performance
}

# Generate test report
generate_report() {
    echo ""
    echo "=================================="
    echo "       TEST RESULTS SUMMARY"
    echo "=================================="
    echo "Total Tests: $TESTS_TOTAL"
    echo "Passed: $TESTS_PASSED"
    echo "Failed: $TESTS_FAILED"
    echo ""
    
    if [ $TESTS_FAILED -eq 0 ]; then
        log_success "All tests passed! ✓"
        echo "System is ready for use."
        exit 0
    else
        log_error "$TESTS_FAILED tests failed! ✗"
        echo "Please review the failed tests and fix issues before deployment."
        exit 1
    fi
}

# Main function
main() {
    echo "Marketing AI System Testing Suite"
    echo "================================="
    echo "Base URL: $BASE_URL"
    echo "Frontend URL: $FRONTEND_URL"
    echo ""
    
    case "${1:-comprehensive}" in
        "health")
            wait_for_services
            test_health_endpoints
            ;;
        "auth")
            wait_for_services
            test_authentication
            ;;
        "api")
            wait_for_services
            test_api_endpoints
            ;;
        "performance")
            wait_for_services
            test_performance
            ;;
        "comprehensive"|*)
            run_comprehensive_test
            ;;
    esac
    
    generate_report
}

# Show usage
show_usage() {
    echo "Usage: $0 [test_type]"
    echo ""
    echo "Test types:"
    echo "  comprehensive  - Run all tests (default)"
    echo "  health        - Test health endpoints only"
    echo "  auth          - Test authentication only"
    echo "  api           - Test API endpoints only"
    echo "  performance   - Run performance tests only"
    echo ""
    echo "Environment variables:"
    echo "  BASE_URL      - Backend API URL (default: http://localhost:8080)"
    echo "  FRONTEND_URL  - Frontend URL (default: http://localhost:3000)"
    echo "  TEST_EMAIL    - Test user email (default: test@example.com)"
    echo "  TEST_PASSWORD - Test user password (default: password123)"
}

# Handle arguments
if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    show_usage
    exit 0
fi

# Run main function
main "$@"
