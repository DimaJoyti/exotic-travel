#!/bin/bash

# Security testing script for Exotic Travel Booking Platform
# This script performs comprehensive security tests including vulnerability scanning and penetration testing

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
TEST_USER_EMAIL="${TEST_USER_EMAIL:-test@example.com}"
TEST_USER_PASSWORD="${TEST_USER_PASSWORD:-TestPassword123!}"

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

print_vulnerability() {
    echo -e "${RED}[VULNERABILITY]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if services are running
check_services() {
    print_status "Checking if services are running..."
    
    # Check backend
    if curl -f -s "$API_BASE_URL/health" > /dev/null 2>&1; then
        print_success "Backend service is running"
    else
        print_error "Backend service is not running at $API_BASE_URL"
        return 1
    fi
    
    return 0
}

# Function to test authentication security
test_authentication_security() {
    print_status "Testing authentication security..."
    
    # Test 1: Check for missing authentication
    print_status "Testing access to protected endpoints without authentication..."
    response=$(curl -s -o /dev/null -w "%{http_code}" "$API_BASE_URL/api/bookings")
    if [ "$response" = "401" ]; then
        print_success "Protected endpoint correctly requires authentication"
    else
        print_vulnerability "Protected endpoint accessible without authentication (HTTP $response)"
    fi
    
    # Test 2: Test weak password acceptance
    print_status "Testing weak password acceptance..."
    weak_passwords=("123456" "password" "admin" "test")
    for weak_pass in "${weak_passwords[@]}"; do
        response=$(curl -s -X POST "$API_BASE_URL/api/auth/register" \
            -H "Content-Type: application/json" \
            -d "{\"email\":\"weak$RANDOM@test.com\",\"password\":\"$weak_pass\",\"name\":\"Test User\"}" \
            -w "%{http_code}" -o /dev/null)
        
        if [ "$response" = "400" ] || [ "$response" = "422" ]; then
            print_success "Weak password '$weak_pass' correctly rejected"
        else
            print_vulnerability "Weak password '$weak_pass' accepted (HTTP $response)"
        fi
    done
    
    # Test 3: Test brute force protection
    print_status "Testing brute force protection..."
    for i in {1..6}; do
        curl -s -X POST "$API_BASE_URL/api/auth/login" \
            -H "Content-Type: application/json" \
            -d "{\"email\":\"nonexistent@test.com\",\"password\":\"wrongpassword\"}" \
            > /dev/null
    done
    
    # Check if rate limiting kicks in
    response=$(curl -s -X POST "$API_BASE_URL/api/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"nonexistent@test.com\",\"password\":\"wrongpassword\"}" \
        -w "%{http_code}" -o /dev/null)
    
    if [ "$response" = "429" ]; then
        print_success "Rate limiting active after multiple failed attempts"
    else
        print_warning "Rate limiting may not be active (HTTP $response)"
    fi
}

# Function to test input validation
test_input_validation() {
    print_status "Testing input validation and sanitization..."
    
    # Test 1: SQL Injection attempts
    print_status "Testing SQL injection protection..."
    sql_payloads=(
        "'; DROP TABLE users; --"
        "' OR '1'='1"
        "' UNION SELECT * FROM users --"
        "admin'--"
        "' OR 1=1#"
    )
    
    for payload in "${sql_payloads[@]}"; do
        response=$(curl -s -X POST "$API_BASE_URL/api/auth/login" \
            -H "Content-Type: application/json" \
            -d "{\"email\":\"$payload\",\"password\":\"test\"}" \
            -w "%{http_code}" -o /dev/null)
        
        if [ "$response" = "400" ] || [ "$response" = "422" ]; then
            print_success "SQL injection payload blocked: $payload"
        else
            print_vulnerability "SQL injection payload may have been processed: $payload (HTTP $response)"
        fi
    done
    
    # Test 2: XSS attempts
    print_status "Testing XSS protection..."
    xss_payloads=(
        "<script>alert('xss')</script>"
        "javascript:alert('xss')"
        "<img src=x onerror=alert('xss')>"
        "<svg onload=alert('xss')>"
        "';alert('xss');//"
    )
    
    for payload in "${xss_payloads[@]}"; do
        response=$(curl -s -X POST "$API_BASE_URL/api/auth/register" \
            -H "Content-Type: application/json" \
            -d "{\"email\":\"test$RANDOM@test.com\",\"password\":\"TestPass123!\",\"name\":\"$payload\"}" \
            -w "%{http_code}" -o /dev/null)
        
        if [ "$response" = "400" ] || [ "$response" = "422" ]; then
            print_success "XSS payload blocked: $payload"
        else
            print_warning "XSS payload may have been accepted: $payload (HTTP $response)"
        fi
    done
    
    # Test 3: Path traversal attempts
    print_status "Testing path traversal protection..."
    path_payloads=(
        "../../../etc/passwd"
        "..\\..\\..\\windows\\system32\\drivers\\etc\\hosts"
        "%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd"
        "....//....//....//etc/passwd"
    )
    
    for payload in "${path_payloads[@]}"; do
        response=$(curl -s "$API_BASE_URL/api/destinations?search=$payload" \
            -w "%{http_code}" -o /dev/null)
        
        if [ "$response" = "400" ] || [ "$response" = "422" ]; then
            print_success "Path traversal payload blocked: $payload"
        else
            print_warning "Path traversal payload may have been processed: $payload (HTTP $response)"
        fi
    done
}

# Function to test security headers
test_security_headers() {
    print_status "Testing security headers..."
    
    # Get headers from a sample endpoint
    headers=$(curl -s -I "$API_BASE_URL/health")
    
    # Check for important security headers
    security_headers=(
        "X-Frame-Options"
        "X-Content-Type-Options"
        "X-XSS-Protection"
        "Strict-Transport-Security"
        "Content-Security-Policy"
        "Referrer-Policy"
    )
    
    for header in "${security_headers[@]}"; do
        if echo "$headers" | grep -qi "$header"; then
            print_success "Security header present: $header"
        else
            print_warning "Security header missing: $header"
        fi
    done
    
    # Check for information disclosure headers
    disclosure_headers=(
        "Server"
        "X-Powered-By"
        "X-AspNet-Version"
    )
    
    for header in "${disclosure_headers[@]}"; do
        if echo "$headers" | grep -qi "$header"; then
            print_warning "Information disclosure header present: $header"
        else
            print_success "Information disclosure header not present: $header"
        fi
    done
}

# Function to test CORS configuration
test_cors_configuration() {
    print_status "Testing CORS configuration..."
    
    # Test 1: Check if CORS headers are present
    response=$(curl -s -H "Origin: http://malicious-site.com" \
        -H "Access-Control-Request-Method: POST" \
        -H "Access-Control-Request-Headers: Content-Type" \
        -X OPTIONS "$API_BASE_URL/api/destinations")
    
    if echo "$response" | grep -q "Access-Control-Allow-Origin"; then
        origin=$(echo "$response" | grep "Access-Control-Allow-Origin" | cut -d: -f2- | tr -d ' \r\n')
        if [ "$origin" = "*" ]; then
            print_vulnerability "CORS allows all origins (*)"
        else
            print_success "CORS has restricted origins: $origin"
        fi
    else
        print_warning "CORS headers not found"
    fi
}

# Function to test SSL/TLS configuration
test_ssl_configuration() {
    print_status "Testing SSL/TLS configuration..."
    
    # Extract hostname and port from URL
    if [[ $API_BASE_URL =~ https://([^:/]+)(:([0-9]+))? ]]; then
        hostname="${BASH_REMATCH[1]}"
        port="${BASH_REMATCH[3]:-443}"
        
        if command_exists openssl; then
            print_status "Testing SSL certificate..."
            ssl_info=$(echo | openssl s_client -connect "$hostname:$port" -servername "$hostname" 2>/dev/null)
            
            if echo "$ssl_info" | grep -q "Verify return code: 0"; then
                print_success "SSL certificate is valid"
            else
                print_warning "SSL certificate validation failed"
            fi
            
            # Check SSL version
            if echo "$ssl_info" | grep -q "TLSv1.2\|TLSv1.3"; then
                print_success "Using secure TLS version"
            else
                print_vulnerability "Using insecure TLS version"
            fi
        else
            print_warning "OpenSSL not available for SSL testing"
        fi
    else
        print_warning "Not using HTTPS - SSL testing skipped"
    fi
}

# Function to test for common vulnerabilities
test_common_vulnerabilities() {
    print_status "Testing for common vulnerabilities..."
    
    # Test 1: Check for debug endpoints
    debug_endpoints=(
        "/debug"
        "/admin"
        "/.env"
        "/config"
        "/swagger"
        "/api-docs"
        "/phpinfo.php"
        "/server-info"
        "/server-status"
    )
    
    for endpoint in "${debug_endpoints[@]}"; do
        response=$(curl -s -o /dev/null -w "%{http_code}" "$API_BASE_URL$endpoint")
        if [ "$response" = "200" ]; then
            print_vulnerability "Debug/admin endpoint accessible: $endpoint"
        fi
    done
    
    # Test 2: Check for HTTP methods
    methods=("TRACE" "TRACK" "DEBUG" "CONNECT")
    for method in "${methods[@]}"; do
        response=$(curl -s -X "$method" -o /dev/null -w "%{http_code}" "$API_BASE_URL/")
        if [ "$response" = "200" ]; then
            print_vulnerability "Dangerous HTTP method allowed: $method"
        fi
    done
    
    # Test 3: Check for directory listing
    response=$(curl -s "$API_BASE_URL/" | grep -i "index of\|directory listing")
    if [ -n "$response" ]; then
        print_vulnerability "Directory listing may be enabled"
    fi
}

# Function to test session management
test_session_management() {
    print_status "Testing session management..."
    
    # Test 1: Check for secure session cookies
    if command_exists curl; then
        # Login and check cookie attributes
        login_response=$(curl -s -c cookies.txt -X POST "$API_BASE_URL/api/auth/login" \
            -H "Content-Type: application/json" \
            -d "{\"email\":\"$TEST_USER_EMAIL\",\"password\":\"$TEST_USER_PASSWORD\"}")
        
        if [ -f cookies.txt ]; then
            if grep -q "HttpOnly" cookies.txt; then
                print_success "Session cookies have HttpOnly flag"
            else
                print_vulnerability "Session cookies missing HttpOnly flag"
            fi
            
            if grep -q "Secure" cookies.txt; then
                print_success "Session cookies have Secure flag"
            else
                print_warning "Session cookies missing Secure flag (may be OK for HTTP)"
            fi
            
            rm -f cookies.txt
        fi
    fi
}

# Function to generate security report
generate_security_report() {
    print_status "Generating security report..."
    
    local report_file="security_report_$(date +%Y%m%d_%H%M%S).txt"
    
    {
        echo "=== Security Test Report ==="
        echo "Generated: $(date)"
        echo "Target: $API_BASE_URL"
        echo ""
        echo "=== Test Summary ==="
        echo "This report contains the results of automated security testing."
        echo "Manual security review is still recommended."
        echo ""
        echo "=== Recommendations ==="
        echo "1. Ensure all security headers are properly configured"
        echo "2. Implement proper input validation and sanitization"
        echo "3. Use HTTPS in production with valid SSL certificates"
        echo "4. Implement rate limiting and brute force protection"
        echo "5. Regular security audits and penetration testing"
        echo "6. Keep all dependencies up to date"
        echo "7. Implement proper logging and monitoring"
        echo ""
        echo "=== Additional Security Measures ==="
        echo "- Web Application Firewall (WAF)"
        echo "- DDoS protection"
        echo "- Regular vulnerability scanning"
        echo "- Security awareness training"
        echo "- Incident response plan"
        
    } > "$report_file"
    
    print_success "Security report saved to: $report_file"
}

# Main function
main() {
    print_status "Starting security testing for Exotic Travel Booking Platform"
    print_status "================================================================"
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --auth-only)
                RUN_AUTH_ONLY=true
                shift
                ;;
            --input-only)
                RUN_INPUT_ONLY=true
                shift
                ;;
            --headers-only)
                RUN_HEADERS_ONLY=true
                shift
                ;;
            --full-scan)
                RUN_FULL_SCAN=true
                shift
                ;;
            --help)
                echo "Usage: $0 [options]"
                echo "Options:"
                echo "  --auth-only      Run only authentication tests"
                echo "  --input-only     Run only input validation tests"
                echo "  --headers-only   Run only security headers tests"
                echo "  --full-scan      Run comprehensive security scan"
                echo "  --help          Show this help message"
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
    if [ "$RUN_AUTH_ONLY" = true ]; then
        test_authentication_security
    elif [ "$RUN_INPUT_ONLY" = true ]; then
        test_input_validation
    elif [ "$RUN_HEADERS_ONLY" = true ]; then
        test_security_headers
    elif [ "$RUN_FULL_SCAN" = true ]; then
        test_authentication_security
        test_input_validation
        test_security_headers
        test_cors_configuration
        test_ssl_configuration
        test_common_vulnerabilities
        test_session_management
    else
        # Run standard security tests
        test_authentication_security
        test_input_validation
        test_security_headers
        test_cors_configuration
    fi
    
    # Generate report
    generate_security_report
    
    print_success "Security testing completed!"
    print_status "================================================================"
    print_warning "Note: This automated testing does not replace manual security review."
    print_warning "Consider hiring security professionals for comprehensive testing."
}

# Run main function with all arguments
main "$@"
