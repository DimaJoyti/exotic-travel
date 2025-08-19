#!/bin/bash

# Deployment script for Exotic Travel Booking Platform
# Usage: ./scripts/deploy.sh [environment] [options]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
ENVIRONMENT="${1:-production}"
DEPLOYMENT_TYPE="${2:-docker-compose}"

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

# Function to validate environment
validate_environment() {
    case $ENVIRONMENT in
        development|staging|production)
            print_status "Deploying to $ENVIRONMENT environment"
            ;;
        *)
            print_error "Invalid environment: $ENVIRONMENT"
            print_error "Valid environments: development, staging, production"
            exit 1
            ;;
    esac
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    case $DEPLOYMENT_TYPE in
        docker-compose)
            if ! command_exists docker; then
                print_error "Docker is not installed"
                exit 1
            fi
            
            if ! command_exists docker-compose; then
                print_error "Docker Compose is not installed"
                exit 1
            fi
            ;;
        kubernetes)
            if ! command_exists kubectl; then
                print_error "kubectl is not installed"
                exit 1
            fi
            
            if ! kubectl cluster-info >/dev/null 2>&1; then
                print_error "Cannot connect to Kubernetes cluster"
                exit 1
            fi
            ;;
        *)
            print_error "Invalid deployment type: $DEPLOYMENT_TYPE"
            print_error "Valid types: docker-compose, kubernetes"
            exit 1
            ;;
    esac
    
    print_success "Prerequisites check passed"
}

# Function to load environment variables
load_environment() {
    local env_file="$PROJECT_ROOT/.env.$ENVIRONMENT"
    
    if [ -f "$env_file" ]; then
        print_status "Loading environment variables from $env_file"
        set -a
        source "$env_file"
        set +a
    else
        print_warning "Environment file $env_file not found"
        if [ "$ENVIRONMENT" = "production" ]; then
            print_error "Production environment file is required"
            exit 1
        fi
    fi
}

# Function to build Docker images
build_images() {
    print_status "Building Docker images..."
    
    cd "$PROJECT_ROOT"
    
    # Build backend image
    print_status "Building backend image..."
    docker build -f backend/Dockerfile.prod -t exotic-travel-backend:latest backend/
    
    # Build frontend image
    print_status "Building frontend image..."
    docker build -f frontend/Dockerfile.prod -t exotic-travel-frontend:latest frontend/
    
    print_success "Docker images built successfully"
}

# Function to deploy with Docker Compose
deploy_docker_compose() {
    print_status "Deploying with Docker Compose..."
    
    cd "$PROJECT_ROOT"
    
    # Create necessary directories
    mkdir -p logs/nginx backups uploads
    
    # Generate SSL certificates if needed
    if [ "$ENVIRONMENT" = "production" ] && [ ! -f "/etc/ssl/certs/fullchain.pem" ]; then
        print_warning "SSL certificates not found. Please set up SSL certificates before production deployment."
    fi
    
    # Deploy services
    docker-compose -f docker-compose.prod.yml down --remove-orphans
    docker-compose -f docker-compose.prod.yml up -d --build
    
    # Wait for services to be ready
    print_status "Waiting for services to be ready..."
    sleep 30
    
    # Run database migrations
    print_status "Running database migrations..."
    docker-compose -f docker-compose.prod.yml exec -T backend go run cmd/migrate/main.go up
    
    print_success "Docker Compose deployment completed"
}

# Function to deploy to Kubernetes
deploy_kubernetes() {
    print_status "Deploying to Kubernetes..."
    
    cd "$PROJECT_ROOT"
    
    # Apply namespace and RBAC
    kubectl apply -f k8s/namespace.yaml
    
    # Apply ConfigMaps and Secrets
    kubectl apply -f k8s/configmap.yaml
    
    # Check if secrets exist, if not, prompt user to create them
    if ! kubectl get secret app-secrets -n exotic-travel >/dev/null 2>&1; then
        print_warning "Secrets not found. Please create secrets before deployment:"
        print_warning "kubectl apply -f k8s/secrets.yaml"
        print_warning "Or use the commands in the secrets.yaml file to create them"
        exit 1
    fi
    
    # Deploy database services
    kubectl apply -f k8s/postgres.yaml
    kubectl apply -f k8s/redis.yaml
    
    # Wait for database services to be ready
    print_status "Waiting for database services..."
    kubectl wait --for=condition=ready pod -l app=postgres -n exotic-travel --timeout=300s
    kubectl wait --for=condition=ready pod -l app=redis -n exotic-travel --timeout=300s
    
    # Deploy application services
    kubectl apply -f k8s/backend.yaml
    kubectl apply -f k8s/frontend.yaml
    
    # Wait for application services to be ready
    print_status "Waiting for application services..."
    kubectl wait --for=condition=ready pod -l app=backend -n exotic-travel --timeout=300s
    kubectl wait --for=condition=ready pod -l app=frontend -n exotic-travel --timeout=300s
    
    # Apply ingress
    kubectl apply -f k8s/ingress.yaml
    
    # Run database migrations
    print_status "Running database migrations..."
    kubectl exec -n exotic-travel deployment/backend -- go run cmd/migrate/main.go up
    
    print_success "Kubernetes deployment completed"
}

# Function to run health checks
run_health_checks() {
    print_status "Running health checks..."
    
    case $DEPLOYMENT_TYPE in
        docker-compose)
            # Check backend health
            if curl -f -s http://localhost:8080/health >/dev/null; then
                print_success "Backend health check passed"
            else
                print_error "Backend health check failed"
                return 1
            fi
            
            # Check frontend health
            if curl -f -s http://localhost:3000/api/health >/dev/null; then
                print_success "Frontend health check passed"
            else
                print_error "Frontend health check failed"
                return 1
            fi
            ;;
        kubernetes)
            # Check backend health
            if kubectl exec -n exotic-travel deployment/backend -- wget -q --spider http://localhost:8080/health; then
                print_success "Backend health check passed"
            else
                print_error "Backend health check failed"
                return 1
            fi
            
            # Check frontend health
            if kubectl exec -n exotic-travel deployment/frontend -- wget -q --spider http://localhost:3000/api/health; then
                print_success "Frontend health check passed"
            else
                print_error "Frontend health check failed"
                return 1
            fi
            ;;
    esac
    
    print_success "All health checks passed"
}

# Function to show deployment status
show_status() {
    print_status "Deployment Status:"
    
    case $DEPLOYMENT_TYPE in
        docker-compose)
            docker-compose -f docker-compose.prod.yml ps
            ;;
        kubernetes)
            kubectl get pods -n exotic-travel
            kubectl get services -n exotic-travel
            kubectl get ingress -n exotic-travel
            ;;
    esac
}

# Function to show logs
show_logs() {
    print_status "Recent logs:"
    
    case $DEPLOYMENT_TYPE in
        docker-compose)
            docker-compose -f docker-compose.prod.yml logs --tail=50
            ;;
        kubernetes)
            kubectl logs -n exotic-travel deployment/backend --tail=50
            kubectl logs -n exotic-travel deployment/frontend --tail=50
            ;;
    esac
}

# Function to rollback deployment
rollback() {
    print_warning "Rolling back deployment..."
    
    case $DEPLOYMENT_TYPE in
        docker-compose)
            docker-compose -f docker-compose.prod.yml down
            print_success "Docker Compose deployment rolled back"
            ;;
        kubernetes)
            kubectl rollout undo deployment/backend -n exotic-travel
            kubectl rollout undo deployment/frontend -n exotic-travel
            print_success "Kubernetes deployment rolled back"
            ;;
    esac
}

# Function to cleanup
cleanup() {
    print_status "Cleaning up..."
    
    case $DEPLOYMENT_TYPE in
        docker-compose)
            docker system prune -f
            ;;
        kubernetes)
            kubectl delete pods --field-selector=status.phase=Succeeded -n exotic-travel
            ;;
    esac
    
    print_success "Cleanup completed"
}

# Main deployment function
main() {
    print_status "Starting deployment of Exotic Travel Booking Platform"
    print_status "Environment: $ENVIRONMENT"
    print_status "Deployment Type: $DEPLOYMENT_TYPE"
    print_status "=================================================="
    
    validate_environment
    check_prerequisites
    load_environment
    
    case $DEPLOYMENT_TYPE in
        docker-compose)
            build_images
            deploy_docker_compose
            ;;
        kubernetes)
            deploy_kubernetes
            ;;
    esac
    
    # Wait a bit for services to stabilize
    sleep 10
    
    if run_health_checks; then
        print_success "Deployment completed successfully!"
        show_status
        
        print_status "Access URLs:"
        case $DEPLOYMENT_TYPE in
            docker-compose)
                print_status "Frontend: http://localhost:3000"
                print_status "Backend API: http://localhost:8080"
                print_status "Monitoring: http://localhost:3001 (Grafana)"
                ;;
            kubernetes)
                print_status "Check ingress for external URLs:"
                kubectl get ingress -n exotic-travel
                ;;
        esac
    else
        print_error "Deployment completed but health checks failed"
        show_logs
        exit 1
    fi
}

# Handle command line arguments
case "${3:-}" in
    --rollback)
        rollback
        exit 0
        ;;
    --status)
        show_status
        exit 0
        ;;
    --logs)
        show_logs
        exit 0
        ;;
    --cleanup)
        cleanup
        exit 0
        ;;
    --help)
        echo "Usage: $0 [environment] [deployment-type] [options]"
        echo ""
        echo "Environments: development, staging, production"
        echo "Deployment Types: docker-compose, kubernetes"
        echo ""
        echo "Options:"
        echo "  --rollback    Rollback the deployment"
        echo "  --status      Show deployment status"
        echo "  --logs        Show recent logs"
        echo "  --cleanup     Clean up resources"
        echo "  --help        Show this help message"
        echo ""
        echo "Examples:"
        echo "  $0 production docker-compose"
        echo "  $0 staging kubernetes"
        echo "  $0 production docker-compose --rollback"
        exit 0
        ;;
esac

# Run main deployment
main
