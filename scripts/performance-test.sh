#!/bin/bash

# Performance testing script for Exotic Travel Booking Platform
# This script runs comprehensive performance tests including load testing, stress testing, and benchmarks

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
FRONTEND_URL="${FRONTEND_URL:-http://localhost:3000}"
CONCURRENT_USERS="${CONCURRENT_USERS:-10}"
TEST_DURATION="${TEST_DURATION:-30s}"
RAMP_UP_TIME="${RAMP_UP_TIME:-10s}"

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

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if services are running
check_services() {
    print_status "Checking if services are running..."
    
    # Check backend
    if curl -f "$API_BASE_URL/health" > /dev/null 2>&1; then
        print_success "Backend service is running"
    else
        print_error "Backend service is not running at $API_BASE_URL"
        return 1
    fi
    
    # Check frontend
    if curl -f "$FRONTEND_URL" > /dev/null 2>&1; then
        print_success "Frontend service is running"
    else
        print_warning "Frontend service is not running at $FRONTEND_URL"
    fi
    
    return 0
}

# Function to run basic load test with curl
run_basic_load_test() {
    print_status "Running basic load test with curl..."
    
    local endpoint="$1"
    local requests="${2:-100}"
    local concurrency="${3:-10}"
    
    if command_exists ab; then
        print_status "Using Apache Bench for load testing..."
        ab -n "$requests" -c "$concurrency" -g "load_test_results.tsv" "$API_BASE_URL$endpoint"
    else
        print_warning "Apache Bench not found. Running simple curl test..."
        
        local start_time=$(date +%s)
        local success_count=0
        local error_count=0
        
        for i in $(seq 1 "$requests"); do
            if curl -f -s "$API_BASE_URL$endpoint" > /dev/null; then
                ((success_count++))
            else
                ((error_count++))
            fi
            
            if [ $((i % 10)) -eq 0 ]; then
                echo -n "."
            fi
        done
        
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        local rps=$((requests / duration))
        
        echo ""
        print_success "Load test completed:"
        print_status "  Total requests: $requests"
        print_status "  Successful: $success_count"
        print_status "  Failed: $error_count"
        print_status "  Duration: ${duration}s"
        print_status "  Requests per second: $rps"
    fi
}

# Function to run API endpoint tests
run_api_tests() {
    print_status "Running API endpoint performance tests..."
    
    # Test health endpoint
    print_status "Testing health endpoint..."
    run_basic_load_test "/health" 50 5
    
    # Test destinations endpoint
    print_status "Testing destinations endpoint..."
    run_basic_load_test "/api/destinations" 30 3
    
    # Test specific destination endpoint
    print_status "Testing specific destination endpoint..."
    run_basic_load_test "/api/destinations/1" 30 3
}

# Function to run database performance tests
run_database_tests() {
    print_status "Running database performance tests..."
    
    # Test database connection
    if command_exists psql; then
        print_status "Testing database connection..."
        
        # Simple query performance test
        local db_host="${DB_HOST:-localhost}"
        local db_port="${DB_PORT:-5432}"
        local db_name="${DB_NAME:-exotic_travel}"
        local db_user="${DB_USER:-postgres}"
        
        local start_time=$(date +%s%N)
        PGPASSWORD="${DB_PASSWORD:-postgres}" psql -h "$db_host" -p "$db_port" -U "$db_user" -d "$db_name" -c "SELECT COUNT(*) FROM destinations;" > /dev/null 2>&1
        local end_time=$(date +%s%N)
        local duration=$(( (end_time - start_time) / 1000000 ))
        
        print_status "Database query took: ${duration}ms"
    else
        print_warning "psql not found. Skipping database tests."
    fi
}

# Function to run memory and CPU tests
run_system_tests() {
    print_status "Running system performance tests..."
    
    # Get current system metrics
    if command_exists free; then
        print_status "Memory usage:"
        free -h
    fi
    
    if command_exists top; then
        print_status "CPU usage (5 second sample):"
        timeout 5 top -b -n1 | head -20
    fi
    
    # Test memory usage during load
    print_status "Monitoring memory during load test..."
    
    # Start memory monitoring in background
    (
        for i in {1..30}; do
            if command_exists ps; then
                ps aux | grep -E "(server|node)" | grep -v grep | awk '{print $6}' | head -5
            fi
            sleep 1
        done
    ) > memory_usage.log &
    
    local monitor_pid=$!
    
    # Run load test
    run_basic_load_test "/health" 100 10
    
    # Stop monitoring
    kill $monitor_pid 2>/dev/null || true
    
    if [ -f memory_usage.log ]; then
        print_status "Memory usage during test (RSS in KB):"
        sort -n memory_usage.log | tail -5
        rm -f memory_usage.log
    fi
}

# Function to run frontend performance tests
run_frontend_tests() {
    print_status "Running frontend performance tests..."
    
    if command_exists curl; then
        # Test static asset loading
        print_status "Testing static asset performance..."
        
        local start_time=$(date +%s%N)
        curl -f -s "$FRONTEND_URL" > /dev/null
        local end_time=$(date +%s%N)
        local duration=$(( (end_time - start_time) / 1000000 ))
        
        print_status "Frontend page load took: ${duration}ms"
        
        # Test multiple concurrent requests
        print_status "Testing concurrent frontend requests..."
        run_basic_load_test "" 20 5 # Empty endpoint for frontend root
    else
        print_warning "curl not found. Skipping frontend tests."
    fi
}

# Function to run cache performance tests
run_cache_tests() {
    print_status "Running cache performance tests..."
    
    # Test Redis if available
    if command_exists redis-cli; then
        print_status "Testing Redis performance..."
        
        local redis_host="${REDIS_HOST:-localhost}"
        local redis_port="${REDIS_PORT:-6379}"
        
        # Simple Redis benchmark
        redis-cli -h "$redis_host" -p "$redis_port" --latency-history -i 1 > redis_latency.log &
        local redis_pid=$!
        
        # Run some cache operations through API
        for i in {1..10}; do
            curl -f -s "$API_BASE_URL/api/destinations/1" > /dev/null
            sleep 0.1
        done
        
        sleep 2
        kill $redis_pid 2>/dev/null || true
        
        if [ -f redis_latency.log ]; then
            print_status "Redis latency (last 5 measurements):"
            tail -5 redis_latency.log
            rm -f redis_latency.log
        fi
    else
        print_warning "redis-cli not found. Skipping cache tests."
    fi
}

# Function to generate performance report
generate_report() {
    print_status "Generating performance report..."
    
    local report_file="performance_report_$(date +%Y%m%d_%H%M%S).txt"
    
    {
        echo "=== Performance Test Report ==="
        echo "Generated: $(date)"
        echo "API Base URL: $API_BASE_URL"
        echo "Frontend URL: $FRONTEND_URL"
        echo "Test Configuration:"
        echo "  Concurrent Users: $CONCURRENT_USERS"
        echo "  Test Duration: $TEST_DURATION"
        echo "  Ramp Up Time: $RAMP_UP_TIME"
        echo ""
        
        # Get current metrics from API if available
        if curl -f -s "$API_BASE_URL/metrics" > /dev/null 2>&1; then
            echo "=== Current API Metrics ==="
            curl -s "$API_BASE_URL/metrics" | head -20
            echo ""
        fi
        
        # System information
        echo "=== System Information ==="
        echo "OS: $(uname -a)"
        echo "CPU: $(nproc) cores"
        
        if command_exists free; then
            echo "Memory:"
            free -h
        fi
        
        if command_exists df; then
            echo "Disk:"
            df -h | head -5
        fi
        
        echo ""
        echo "=== Test Results ==="
        echo "See individual test outputs above for detailed results."
        
    } > "$report_file"
    
    print_success "Performance report saved to: $report_file"
}

# Function to run stress test
run_stress_test() {
    print_status "Running stress test..."
    
    if command_exists ab; then
        print_status "Running high-load stress test..."
        
        # Gradually increase load
        for concurrent in 5 10 20 50; do
            print_status "Testing with $concurrent concurrent users..."
            ab -n 100 -c "$concurrent" "$API_BASE_URL/health" | grep -E "(Requests per second|Time per request|Transfer rate)"
            sleep 2
        done
    else
        print_warning "Apache Bench not found. Running basic stress test..."
        
        # Simple stress test with curl
        print_status "Running concurrent curl requests..."
        
        for i in {1..50}; do
            curl -f -s "$API_BASE_URL/health" > /dev/null &
        done
        
        wait
        print_success "Stress test completed"
    fi
}

# Main function
main() {
    print_status "Starting performance testing for Exotic Travel Booking Platform"
    print_status "=============================================================="
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --api-only)
                RUN_API_ONLY=true
                shift
                ;;
            --frontend-only)
                RUN_FRONTEND_ONLY=true
                shift
                ;;
            --stress-test)
                RUN_STRESS_TEST=true
                shift
                ;;
            --concurrent-users)
                CONCURRENT_USERS="$2"
                shift 2
                ;;
            --duration)
                TEST_DURATION="$2"
                shift 2
                ;;
            --help)
                echo "Usage: $0 [options]"
                echo "Options:"
                echo "  --api-only           Run only API tests"
                echo "  --frontend-only      Run only frontend tests"
                echo "  --stress-test        Include stress testing"
                echo "  --concurrent-users N Set number of concurrent users"
                echo "  --duration TIME      Set test duration (e.g., 30s, 2m)"
                echo "  --help              Show this help message"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Check if services are running
    if ! check_services; then
        print_error "Services are not running. Please start the services first."
        exit 1
    fi
    
    # Run tests based on flags
    if [ "$RUN_API_ONLY" = true ]; then
        run_api_tests
        run_database_tests
        run_cache_tests
    elif [ "$RUN_FRONTEND_ONLY" = true ]; then
        run_frontend_tests
    else
        # Run all tests
        run_api_tests
        run_database_tests
        run_cache_tests
        run_frontend_tests
        run_system_tests
    fi
    
    if [ "$RUN_STRESS_TEST" = true ]; then
        run_stress_test
    fi
    
    # Generate report
    generate_report
    
    print_success "Performance testing completed successfully!"
    print_status "=============================================================="
}

# Run main function with all arguments
main "$@"
