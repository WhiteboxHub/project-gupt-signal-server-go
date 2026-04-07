# Quick Start Guide

Get your signaling server running in **5 minutes**!

## Prerequisites

- Go 1.22+ installed
- Terminal access

## Step 1: Clone & Setup (30 seconds)

```bash
cd /path/to/signaling-server
go mod download
```

## Step 2: Build (10 seconds)

```bash
go build -o bin/signaling-server ./cmd/server
```

## Step 3: Run Server (5 seconds)

```bash
# Option A: Run directly
./bin/signaling-server

# Option B: Using go run
go run ./cmd/server/main.go

# Option C: Using Makefile
make run
```

**Server is now running on http://localhost:8080** ✅

## Step 4: Test It (2 minutes)

### Terminal 1: Start Server
```bash
make run
```

You should see:
```
=== Signaling Server ===
Configuration:
  Address: :8080
  Session TTL: 1h0m0s
  Monitor Interval: 30s
========================
[Server] Starting on :8080
```

### Terminal 2: Run Host
```bash
make run-host
```

You should see:
```
=== HOST TEST CLIENT ===
Connected to server
Sent register message (session=TEST123)
Waiting for client to connect...
```

### Terminal 3: Run Client
```bash
make run-client
```

You should see:
```
=== CLIENT TEST CLIENT ===
Connected to server
Sent connect message (session=TEST123)
Received: type=peer_info
=== PEER INFORMATION ===
Host Local IP:  192.168.1.100:9000
Host Public IP: 127.0.0.1:54321
========================
```

**Both peers have exchanged information!** 🎉

## What Just Happened?

1. **Host** registered a session (ID: TEST123)
2. **Client** joined that session
3. **Server** exchanged peer information
4. Both peers now have each other's:
   - Local IP and port (for LAN connections)
   - Public IP and port (for NAT traversal)

## Next Steps

### Try with Custom Session

**Host:**
```bash
go run ./cmd/testclient/host.go -session=MYSESSION -password=secret123
```

**Client:**
```bash
go run ./cmd/testclient/client.go -session=MYSESSION -password=secret123
```

### Test Health Check

```bash
curl http://localhost:8080/health
# Response: {"status":"healthy"}
```

### Run with Docker

```bash
# Build image
docker build -t signaling-server .

# Run container
docker run -p 8080:8080 signaling-server
```

### Deploy to Google Cloud

```bash
# Set project
export GCP_PROJECT=your-project-id

# Deploy
./scripts/deploy.sh
```

## Troubleshooting

### Port Already in Use?

```bash
# Kill process on port 8080
lsof -ti:8080 | xargs kill -9
```

### Build Fails?

```bash
# Update dependencies
go mod tidy
go mod download
```

### Connection Refused?

- Make sure server is running
- Check firewall settings
- Verify port 8080 is not blocked

## Understanding the Flow

```
1. Host connects to server
   ↓
2. Host registers session with ID
   ↓
3. Server acknowledges registration
   ↓
4. Client connects to server
   ↓
5. Client joins session with same ID
   ↓
6. Server validates password (if set)
   ↓
7. Server sends host info to client
   ↓
8. Server sends client info to host
   ↓
9. Both peers have each other's details
   ↓
10. Peers can now connect directly (P2P)
```

## Message Examples

### Register Host
```json
{
  "type": "register",
  "session_id": "ABC123",
  "password": "secret",
  "local_ip": "192.168.1.100",
  "local_port": 9000
}
```

### Connect Client
```json
{
  "type": "connect",
  "session_id": "ABC123",
  "password": "secret",
  "local_ip": "192.168.1.200",
  "local_port": 9001
}
```

### Server Response (Peer Info)
```json
{
  "type": "peer_info",
  "peer": {
    "local_ip": "192.168.1.100",
    "local_port": 9000,
    "public_ip": "203.0.113.1",
    "public_port": 54321
  }
}
```

## Configuration Options

```bash
# Custom port
go run ./cmd/server/main.go -addr=:9090

# Custom session TTL (2 hours)
go run ./cmd/server/main.go -session-ttl=2h

# Custom monitor interval (1 minute)
go run ./cmd/server/main.go -monitor-interval=1m

# All together
go run ./cmd/server/main.go \
  -addr=:9090 \
  -session-ttl=2h \
  -monitor-interval=1m
```

## Project Structure

```
signaling-server/
├── cmd/
│   ├── server/          # Main server application
│   └── testclient/      # Test clients (host & client)
├── internal/
│   ├── server/          # WebSocket server
│   ├── connection/      # Connection management
│   ├── session/         # Session management
│   ├── handler/         # Message handling
│   ├── nat/             # NAT traversal helpers
│   └── health/          # Health monitoring
├── Makefile             # Build commands
├── Dockerfile           # Docker configuration
└── README.md            # Full documentation
```

## Available Commands

```bash
make help          # Show all commands
make build         # Build binary
make run           # Run server
make run-host      # Run host test client
make run-client    # Run client test client
make test          # Run tests
make docker-build  # Build Docker image
make docker-run    # Run Docker container
make deploy-gcp    # Deploy to Google Cloud
```

## Learn More

- **Full Documentation**: [README.md](README.md)
- **Architecture Details**: [ARCHITECTURE.md](ARCHITECTURE.md)
- **Deployment Guide**: [DEPLOYMENT.md](DEPLOYMENT.md)
- **Testing Guide**: [TESTING.md](TESTING.md)

## Support

Questions? Issues?

1. Check the documentation files above
2. Review message examples in `examples/`
3. Run test scripts in `scripts/`
4. Open an issue on GitHub

---

**You're all set!** The signaling server is now coordinating P2P connections. 🚀
