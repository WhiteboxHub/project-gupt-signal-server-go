# Testing Guide

## Overview

This guide covers testing the signaling server at various levels: unit tests, integration tests, manual testing, and load testing.

## Unit Tests

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package
go test -v ./internal/session

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Test Coverage by Package

**Session Manager** (`internal/session/`)
- Session creation
- Session joining
- Password validation
- Session expiry
- Concurrent access

**Connection Manager** (`internal/connection/`)
- Connection registration
- Message sending
- Connection removal
- Concurrent operations

**Message Handler** (`internal/handler/`)
- Message parsing
- Message routing
- Error handling
- Peer exchange

**NAT Helper** (`internal/nat/`)
- IP extraction
- Private IP detection
- Port parsing

## Integration Tests

### Manual Integration Test

**Step 1: Start Server**
```bash
# Terminal 1
go run ./cmd/server/main.go
```

Expected output:
```
=== Signaling Server ===
Configuration:
  Address: :8080
  Session TTL: 1h0m0s
  Monitor Interval: 30s
========================
[Server] Starting on :8080
```

**Step 2: Register Host**
```bash
# Terminal 2
go run ./cmd/testclient/host.go \
  -server=ws://localhost:8080/ws \
  -session=TEST123 \
  -password=secret \
  -local-ip=192.168.1.100 \
  -local-port=9000
```

Expected output:
```
=== HOST TEST CLIENT ===
Server: ws://localhost:8080/ws
Session: TEST123
========================
Connected to server
Sent register message (session=TEST123)
Waiting for client to connect...
Received: type=registered
```

**Step 3: Connect Client**
```bash
# Terminal 3
go run ./cmd/testclient/client.go \
  -server=ws://localhost:8080/ws \
  -session=TEST123 \
  -password=secret \
  -local-ip=192.168.1.200 \
  -local-port=9001
```

Expected output:
```
=== CLIENT TEST CLIENT ===
Server: ws://localhost:8080/ws
Session: TEST123
==========================
Connected to server
Sent connect message (session=TEST123)
Received: type=peer_info
=== PEER INFORMATION ===
Host Local IP:  192.168.1.100:9000
Host Public IP: 127.0.0.1:54321
========================
```

**Step 4: Verify Peer Exchange**

Both host and client should receive peer_info messages with:
- Local IP and port (as provided)
- Public IP and port (extracted by server)

### Test Scenarios

#### 1. Successful Connection
✅ Host registers → Client connects → Peer info exchanged

#### 2. Invalid Session
```bash
# Client tries to connect to non-existent session
go run ./cmd/testclient/client.go -session=INVALID
```
Expected: Error message "session not found"

#### 3. Wrong Password
```bash
# Host
go run ./cmd/testclient/host.go -session=TEST -password=correct

# Client
go run ./cmd/testclient/client.go -session=TEST -password=wrong
```
Expected: Error message "invalid password"

#### 4. Duplicate Session
```bash
# Host 1
go run ./cmd/testclient/host.go -session=DUP

# Host 2 (in another terminal)
go run ./cmd/testclient/host.go -session=DUP
```
Expected: Error message "session already exists"

#### 5. Session Full
```bash
# Host
go run ./cmd/testclient/host.go -session=FULL

# Client 1
go run ./cmd/testclient/client.go -session=FULL

# Client 2 (should fail)
go run ./cmd/testclient/client.go -session=FULL
```
Expected: Error message "session already has a client"

#### 6. Heartbeat
Test clients automatically send heartbeats every 30 seconds. Server should respond with heartbeat acknowledgment.

#### 7. Connection Drop
Kill a client (Ctrl+C). Server should detect disconnection and clean up the session.

#### 8. Session Expiry
Create a session and wait for TTL (default 1 hour). Monitor logs for cleanup messages.

For faster testing:
```bash
go run ./cmd/server/main.go -session-ttl=2m
```

## Load Testing

### Using wscat

```bash
# Install wscat
npm install -g wscat

# Connect to server
wscat -c ws://localhost:8080/ws

# Send messages manually
> {"type":"register","session_id":"LOAD1","local_ip":"192.168.1.1","local_port":9000}
```

### Simple Load Test Script

Create `scripts/load_test.sh`:

```bash
#!/bin/bash

# Number of concurrent connections
CONNECTIONS=100
SERVER="ws://localhost:8080/ws"

echo "Starting load test with $CONNECTIONS connections..."

for i in $(seq 1 $CONNECTIONS); do
  (
    go run ./cmd/testclient/host.go \
      -server=$SERVER \
      -session=LOAD$i &
  )
done

wait
echo "Load test complete"
```

Run:
```bash
chmod +x scripts/load_test.sh
./scripts/load_test.sh
```

### Monitor Server Performance

While running load tests, monitor:

**Server Logs:**
```bash
# Watch connection count
tail -f server.log | grep "Stats:"
```

**System Resources:**
```bash
# CPU and memory usage
top -pid $(pgrep signaling-server)

# Connection count
lsof -i :8080 | wc -l
```

## Testing with curl

### Health Check
```bash
curl http://localhost:8080/health
# Expected: {"status":"healthy"}
```

### WebSocket Connection (using websocat)
```bash
# Install websocat
brew install websocat  # macOS
# or cargo install websocat

# Connect and send messages
echo '{"type":"register","session_id":"CURL1","local_ip":"192.168.1.1","local_port":9000}' | \
  websocat ws://localhost:8080/ws
```

## Docker Testing

### Build and Test

```bash
# Build image
docker build -t signaling-server:test .

# Run container
docker run -p 8080:8080 signaling-server:test

# Test connection
go run ./cmd/testclient/host.go -server=ws://localhost:8080/ws
```

### Test Health Endpoint

```bash
docker run -d -p 8080:8080 --name test-server signaling-server:test

# Wait for startup
sleep 2

# Test health
curl http://localhost:8080/health

# Test WebSocket
go run ./cmd/testclient/host.go

# Cleanup
docker stop test-server
docker rm test-server
```

## Cloud Run Testing

### Deploy Test Instance

```bash
# Deploy to Cloud Run
gcloud run deploy signaling-server-test \
  --image gcr.io/$GCP_PROJECT/signaling-server \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated

# Get URL
export TEST_URL=$(gcloud run services describe signaling-server-test \
  --platform managed \
  --region us-central1 \
  --format 'value(status.url)')

# Convert to WebSocket URL
export WS_URL=${TEST_URL/https/wss}/ws

# Test connection
go run ./cmd/testclient/host.go -server=$WS_URL -session=CLOUD1
go run ./cmd/testclient/client.go -server=$WS_URL -session=CLOUD1
```

### Cleanup Test Instance

```bash
gcloud run services delete signaling-server-test \
  --platform managed \
  --region us-central1
```

## Performance Benchmarks

### Connection Latency

```bash
# Measure connection establishment time
time go run ./cmd/testclient/host.go -session=PERF1
```

### Message Round-Trip Time

Modify test client to measure:
- Time to send register message
- Time to receive registered response
- Time to receive peer_info

### Throughput

Create a benchmark that:
1. Opens N connections
2. Sends M messages per connection
3. Measures total time

Example target: 1000 connections, 10 messages each in <5 seconds

## Automated Testing

### GitHub Actions (Example)

Create `.github/workflows/test.yml`:

```yaml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: Download dependencies
        run: go mod download

      - name: Run tests
        run: go test -v ./...

      - name: Run integration test
        run: |
          go run ./cmd/server/main.go &
          sleep 2
          go run ./cmd/testclient/host.go -session=CI1 &
          sleep 1
          go run ./cmd/testclient/client.go -session=CI1
```

## Debugging

### Enable Debug Logging

Modify `cmd/server/main.go`:
```go
log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile | log.Lmicroseconds)
```

### Trace WebSocket Messages

Add logging in `internal/server/websocket.go`:
```go
func (s *Server) handleConnection(connID, remoteAddr string, conn *websocket.Conn) {
    // ...
    for {
        messageType, data, err := conn.ReadMessage()
        log.Printf("[DEBUG] Received raw message: %s", string(data))
        // ...
    }
}
```

### Use Wireshark

Capture WebSocket traffic:
```bash
# Capture on loopback interface
sudo tcpdump -i lo0 -w capture.pcap port 8080

# Analyze in Wireshark
wireshark capture.pcap
```

## Troubleshooting

### Server Won't Start
```bash
# Check if port is in use
lsof -i :8080

# Kill process
kill -9 $(lsof -ti:8080)
```

### Client Can't Connect
```bash
# Check server is running
curl http://localhost:8080/health

# Check firewall
sudo pfctl -sr  # macOS
sudo ufw status  # Linux
```

### Messages Not Being Received
- Check message format (must be valid JSON)
- Verify session_id matches
- Check server logs for errors
- Ensure WebSocket connection is open

### Performance Issues
```bash
# Check system limits
ulimit -n  # Max open files

# Increase if needed
ulimit -n 10000
```

## Test Checklist

Before deploying to production:

- [ ] All unit tests pass
- [ ] Integration test successful (host + client)
- [ ] Invalid session handled
- [ ] Wrong password rejected
- [ ] Duplicate session prevented
- [ ] Session full error works
- [ ] Heartbeat mechanism works
- [ ] Connection cleanup works
- [ ] Session expiry works
- [ ] Load test with 100+ connections
- [ ] Docker build successful
- [ ] Docker container runs correctly
- [ ] Cloud Run deployment works
- [ ] Health endpoint responds
- [ ] Performance meets targets
- [ ] Memory usage acceptable
- [ ] No goroutine leaks

## Continuous Monitoring

In production, monitor:

1. **Connection metrics**: Total connections, connection rate
2. **Session metrics**: Active sessions, session duration
3. **Error rate**: Failed connections, invalid messages
4. **Performance**: Message latency, CPU, memory
5. **Availability**: Uptime, health check status

Use Cloud Monitoring, Prometheus, or similar tools.
