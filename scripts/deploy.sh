#!/bin/bash

# Deployment script for Google Cloud Run
# Usage: ./scripts/deploy.sh [project-id] [region]

set -e

# Default values
DEFAULT_REGION="us-central1"

# Get parameters
GCP_PROJECT=${1:-$GCP_PROJECT}
GCP_REGION=${2:-$DEFAULT_REGION}

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_info() {
    echo -e "${YELLOW}➜ $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# Validate project ID
if [ -z "$GCP_PROJECT" ]; then
    echo "Error: GCP_PROJECT not set"
    echo "Usage: ./scripts/deploy.sh [project-id] [region]"
    echo "Or set GCP_PROJECT environment variable"
    exit 1
fi

echo "=== Deploying to Google Cloud Run ==="
echo "Project: $GCP_PROJECT"
echo "Region: $GCP_REGION"
echo ""

# Set project
print_info "Setting GCP project..."
gcloud config set project $GCP_PROJECT

# Build and push image
print_info "Building and pushing Docker image..."
gcloud builds submit --tag gcr.io/$GCP_PROJECT/signaling-server

print_success "Image built and pushed"

# Deploy to Cloud Run
print_info "Deploying to Cloud Run..."
gcloud run deploy signaling-server \
  --image gcr.io/$GCP_PROJECT/signaling-server \
  --platform managed \
  --region $GCP_REGION \
  --allow-unauthenticated \
  --port 8080 \
  --memory 512Mi \
  --cpu 1 \
  --max-instances 10 \
  --timeout 3600

print_success "Deployment complete"

# Get service URL
print_info "Getting service URL..."
SERVICE_URL=$(gcloud run services describe signaling-server \
  --platform managed \
  --region $GCP_REGION \
  --format 'value(status.url)')

echo ""
print_success "Signaling server deployed successfully!"
echo ""
echo "Service URL: $SERVICE_URL"
echo "Health check: $SERVICE_URL/health"
echo "WebSocket: ${SERVICE_URL/https/wss}/ws"
echo ""
echo "Test with:"
echo "go run ./cmd/testclient/host.go -server=${SERVICE_URL/https/wss}/ws"
