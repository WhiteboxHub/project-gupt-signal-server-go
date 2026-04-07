# Deployment Guide

## Local Development Setup

### Prerequisites

#### macOS
```bash
# Install Homebrew (if not already installed)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install Go
brew install go

# Verify installation
go version  # Should show Go 1.22 or later
```

#### Windows
```powershell
# Download and install Go from: https://go.dev/dl/
# Or use Chocolatey:
choco install golang

# Verify installation
go version
```

### Initialize Project

```bash
# Navigate to project directory
cd /path/to/signaling-server

# Download dependencies
go mod download
go mod tidy

# Verify build
go build ./cmd/server
```

### Run Server Locally

```bash
# Option 1: Using go run
go run ./cmd/server/main.go

# Option 2: Using Makefile
make run

# Option 3: Build and run binary
make build
./bin/signaling-server

# With custom configuration
go run ./cmd/server/main.go -addr=:9090 -session-ttl=2h
```

The server will start on `http://localhost:8080`

### Testing with Test Clients

#### Terminal 1: Start Server
```bash
make run
```

#### Terminal 2: Start Host
```bash
# Basic usage
make run-host

# With custom parameters
go run ./cmd/testclient/host.go \
  -server=ws://localhost:8080/ws \
  -session=ABC123 \
  -password=secret \
  -local-ip=192.168.1.100 \
  -local-port=9000
```

#### Terminal 3: Start Client
```bash
# Basic usage
make run-client

# With custom parameters
go run ./cmd/testclient/client.go \
  -server=ws://localhost:8080/ws \
  -session=ABC123 \
  -password=secret \
  -local-ip=192.168.1.200 \
  -local-port=9001
```

### Expected Output

**Server:**
```
2026/04/06 22:00:00 === Signaling Server ===
2026/04/06 22:00:00 Configuration:
2026/04/06 22:00:00   Address: :8080
2026/04/06 22:00:00   Session TTL: 1h0m0s
2026/04/06 22:00:00   Monitor Interval: 30s
2026/04/06 22:00:00 [Server] Starting on :8080
2026/04/06 22:00:05 [Server] New WebSocket connection: <conn-id> from 127.0.0.1:54321
2026/04/06 22:00:05 [Handler] Host registered session=ABC123
2026/04/06 22:00:10 [Server] New WebSocket connection: <conn-id> from 127.0.0.1:54322
2026/04/06 22:00:10 [Handler] Client joined session=ABC123
2026/04/06 22:00:10 [Handler] Peer exchange completed for session=ABC123
```

**Host Client:**
```
=== HOST TEST CLIENT ===
Server: ws://localhost:8080/ws
Session: ABC123
========================
Connected to server
Sent register message (session=ABC123)
Waiting for client to connect...
Received: type=registered
Received: type=peer_info
=== PEER INFORMATION ===
Client Local IP:  192.168.1.200:9001
Client Public IP: 127.0.0.1:54322
========================
You can now establish P2P connection using these details!
```

**Client:**
```
=== CLIENT TEST CLIENT ===
Server: ws://localhost:8080/ws
Session: ABC123
==========================
Connected to server
Sent connect message (session=ABC123)
Received: type=peer_info
=== PEER INFORMATION ===
Host Local IP:  192.168.1.100:9000
Host Public IP: 127.0.0.1:54321
========================
You can now establish P2P connection using these details!
```

---

## Docker Deployment

### Build Docker Image

```bash
# Build image
docker build -t signaling-server:latest .

# Verify image
docker images | grep signaling-server
```

### Run Container Locally

```bash
# Run container
docker run -p 8080:8080 signaling-server:latest

# Run with custom configuration
docker run -p 9090:8080 signaling-server:latest ./signaling-server -addr=:8080

# Run in background
docker run -d -p 8080:8080 --name signaling signaling-server:latest

# View logs
docker logs -f signaling

# Stop container
docker stop signaling
docker rm signaling
```

### Test Docker Container

```bash
# Health check
curl http://localhost:8080/health

# WebSocket test using test clients
go run ./cmd/testclient/host.go -server=ws://localhost:8080/ws
```

---

## Google Cloud Platform Deployment

### Prerequisites

#### Install gcloud CLI

**macOS:**
```bash
brew install google-cloud-sdk
```

**Windows:**
Download from: https://cloud.google.com/sdk/docs/install

**Linux:**
```bash
curl https://sdk.cloud.google.com | bash
exec -l $SHELL
```

#### Initialize gcloud

```bash
# Login to Google Cloud
gcloud auth login

# Set default project (optional)
gcloud config set project YOUR_PROJECT_ID
```

### Create GCP Project

```bash
# Create new project
gcloud projects create signaling-server-prod --name="Signaling Server"

# Set as active project
gcloud config set project signaling-server-prod

# Enable billing (required for Cloud Run)
# Go to: https://console.cloud.google.com/billing
```

### Enable Required Services

```bash
# Enable Cloud Run API
gcloud services enable run.googleapis.com

# Enable Container Registry
gcloud services enable containerregistry.googleapis.com

# Enable Cloud Build (for building images)
gcloud services enable cloudbuild.googleapis.com
```

### Deploy to Cloud Run

#### Method 1: Using gcloud (Recommended)

```bash
# Set environment variables
export GCP_PROJECT=signaling-server-prod
export GCP_REGION=us-central1

# Build and push image
gcloud builds submit --tag gcr.io/$GCP_PROJECT/signaling-server

# Deploy to Cloud Run
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

# Get service URL
gcloud run services describe signaling-server \
  --platform managed \
  --region $GCP_REGION \
  --format 'value(status.url)'
```

#### Method 2: Using Makefile

```bash
# Update Makefile with your project ID
export GCP_PROJECT=signaling-server-prod

# Deploy
make deploy-gcp
```

### Configure Custom Domain (Optional)

```bash
# Map custom domain
gcloud run domain-mappings create \
  --service signaling-server \
  --domain signal.yourdomain.com \
  --region $GCP_REGION

# Get DNS records to configure
gcloud run domain-mappings describe \
  --domain signal.yourdomain.com \
  --region $GCP_REGION
```

### Environment Variables

```bash
# Deploy with environment variables
gcloud run deploy signaling-server \
  --image gcr.io/$GCP_PROJECT/signaling-server \
  --platform managed \
  --region $GCP_REGION \
  --set-env-vars "ADDR=:8080,SESSION_TTL=3600"
```

### View Logs

```bash
# Stream logs
gcloud run logs tail signaling-server --project=$GCP_PROJECT

# View logs in Cloud Console
gcloud run services describe signaling-server \
  --platform managed \
  --region $GCP_REGION \
  --format 'value(status.url)' | sed 's|https://|https://console.cloud.google.com/run/detail/'$GCP_REGION'/|'
```

### Update Deployment

```bash
# Rebuild and deploy new version
gcloud builds submit --tag gcr.io/$GCP_PROJECT/signaling-server
gcloud run deploy signaling-server \
  --image gcr.io/$GCP_PROJECT/signaling-server \
  --platform managed \
  --region $GCP_REGION
```

### Monitoring & Scaling

```bash
# Set scaling parameters
gcloud run services update signaling-server \
  --min-instances 1 \
  --max-instances 20 \
  --concurrency 1000 \
  --cpu 2 \
  --memory 1Gi \
  --region $GCP_REGION

# View metrics
gcloud run services describe signaling-server \
  --platform managed \
  --region $GCP_REGION
```

---

## Alternative: Compute Engine Deployment

### Create VM Instance

```bash
# Create VM
gcloud compute instances create signaling-server \
  --zone us-central1-a \
  --machine-type e2-micro \
  --image-family ubuntu-2204-lts \
  --image-project ubuntu-os-cloud \
  --boot-disk-size 10GB \
  --tags http-server,https-server

# Configure firewall
gcloud compute firewall-rules create allow-signaling \
  --allow tcp:8080 \
  --source-ranges 0.0.0.0/0 \
  --target-tags http-server

# SSH into instance
gcloud compute ssh signaling-server --zone us-central1-a
```

### Install Dependencies on VM

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Go
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install Docker (optional)
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Clone repository
git clone <your-repo-url>
cd signaling-server

# Build and run
go build -o signaling-server ./cmd/server
./signaling-server
```

### Setup as Systemd Service

```bash
# Create service file
sudo nano /etc/systemd/system/signaling-server.service
```

Content:
```ini
[Unit]
Description=Signaling Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/root/signaling-server
ExecStart=/root/signaling-server/signaling-server
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable signaling-server
sudo systemctl start signaling-server

# Check status
sudo systemctl status signaling-server

# View logs
sudo journalctl -u signaling-server -f
```

---

## Production Considerations

### Security

#### 1. Enable TLS/WSS

```bash
# Using Cloud Run (automatic HTTPS)
# Cloud Run automatically provides TLS

# For Compute Engine, use Let's Encrypt
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d signal.yourdomain.com
```

#### 2. Restrict Origins

Update `internal/server/websocket.go`:
```go
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        origin := r.Header.Get("Origin")
        allowedOrigins := []string{
            "https://yourdomain.com",
            "https://app.yourdomain.com",
        }
        for _, allowed := range allowedOrigins {
            if origin == allowed {
                return true
            }
        }
        return false
    },
}
```

#### 3. Add Authentication

Implement JWT or API key authentication in message handler.

### Monitoring

#### Cloud Monitoring (GCP)

```bash
# Enable monitoring
gcloud services enable monitoring.googleapis.com

# Create uptime check
gcloud monitoring uptime-checks create https signaling-server-check \
  --resource-type uptime-url \
  --hostname YOUR_CLOUD_RUN_URL \
  --path /health
```

#### Prometheus Metrics (Optional)

Add metrics endpoint:
```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

http.Handle("/metrics", promhttp.Handler())
```

### Cost Optimization

**Cloud Run Pricing (as of 2026):**
- First 2 million requests/month: FREE
- CPU: $0.00002400 per vCPU-second
- Memory: $0.00000250 per GiB-second
- Requests: $0.40 per million requests

**Recommendations:**
- Use `--min-instances=0` for low traffic
- Set appropriate `--max-instances` to control costs
- Monitor usage in Cloud Console

---

## Testing Production Deployment

### Test Cloud Run Deployment

```bash
# Get service URL
export SERVICE_URL=$(gcloud run services describe signaling-server \
  --platform managed \
  --region us-central1 \
  --format 'value(status.url)')

# Replace https:// with wss://
export WS_URL=${SERVICE_URL/https/wss}/ws

# Test with clients
go run ./cmd/testclient/host.go -server=$WS_URL -session=PROD123
go run ./cmd/testclient/client.go -server=$WS_URL -session=PROD123
```

### Load Testing

```bash
# Install wscat for WebSocket testing
npm install -g wscat

# Or use k6 for load testing
brew install k6
```

---

## Troubleshooting

### Common Issues

**1. Port already in use**
```bash
# Kill process on port 8080
lsof -ti:8080 | xargs kill -9
```

**2. Cloud Run timeout**
```bash
# Increase timeout
gcloud run services update signaling-server --timeout=3600
```

**3. Connection refused**
```bash
# Check firewall rules
gcloud compute firewall-rules list

# Add rule if needed
gcloud compute firewall-rules create allow-8080 --allow tcp:8080
```

**4. WebSocket upgrade failed**
- Ensure Cloud Run allows WebSocket connections (enabled by default)
- Check CORS/Origin settings

---

## Cleanup

### Delete Cloud Run Service

```bash
gcloud run services delete signaling-server \
  --platform managed \
  --region us-central1
```

### Delete GCP Project

```bash
gcloud projects delete signaling-server-prod
```

### Stop Local Docker

```bash
docker stop signaling
docker rm signaling
docker rmi signaling-server:latest
```
