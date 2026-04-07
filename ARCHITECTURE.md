# Signaling Server Architecture

## Overview
Production-ready WebSocket-based signaling server for P2P remote desktop connections with NAT traversal support.

## System Components

### 1. WebSocket Server (`internal/server/websocket.go`)
- Handles HTTP → WebSocket upgrade
- Manages concurrent connections
- Graceful shutdown support
- Port: 8080 (configurable)

### 2. Connection Manager (`internal/connection/manager.go`)
- Tracks all active WebSocket connections
- Maps connection ID to WebSocket connection
- Thread-safe operations using sync.RWMutex
- Connection lifecycle management

### 3. Session Manager (`internal/session/manager.go`)
- In-memory session storage (sync.Map for concurrency)
- Session lifecycle: create → active → expired
- Password validation
- Session cleanup (TTL: 1 hour)

### 4. Message Handler (`internal/handler/message.go`)
- Routes messages by type
- Validates message format
- Coordinates peer exchange
- Error handling & responses

### 5. NAT Helper (`internal/nat/helper.go`)
- Extracts public IP from WebSocket connection
- Stores both local and public IPs
- Enables UDP hole punching coordination

### 6. Heartbeat Monitor (`internal/health/heartbeat.go`)
- Periodic ping/pong to detect dead connections
- Configurable interval (30s default)
- Auto-cleanup of stale sessions

## Message Protocol

### Message Types

```json
// Register Host
{
  "type": "register",
  "session_id": "ABC123",
  "password": "optional_password",
  "local_ip": "192.168.1.100",
  "local_port": 9000
}

// Connect Client
{
  "type": "connect",
  "session_id": "ABC123",
  "password": "optional_password",
  "local_ip": "192.168.1.200",
  "local_port": 9001
}

// Peer Info Exchange
{
  "type": "peer_info",
  "peer": {
    "local_ip": "192.168.1.100",
    "local_port": 9000,
    "public_ip": "203.0.113.1",
    "public_port": 12345
  }
}

// Heartbeat
{
  "type": "heartbeat"
}

// Error Response
{
  "type": "error",
  "message": "Session not found"
}
```

## Concurrency Model

- **Goroutines**: One per WebSocket connection for reading messages
- **Channels**: Used for graceful shutdown coordination
- **sync.Map**: Lock-free session storage
- **sync.RWMutex**: Connection manager protection

## Data Structures

### Session
```go
type Session struct {
    ID           string
    Password     string
    Host         *PeerInfo
    Client       *PeerInfo
    HostConnID   string
    ClientConnID string
    CreatedAt    time.Time
    Status       SessionStatus
}
```

### PeerInfo
```go
type PeerInfo struct {
    LocalIP    string
    LocalPort  int
    PublicIP   string
    PublicPort int
}
```

## Scalability Considerations

### Current (In-Memory)
- Single server instance
- Up to ~10K concurrent connections
- Session data in memory (lost on restart)

### Production Enhancements
- **Redis**: Distributed session storage
- **Load Balancer**: Multiple server instances
- **Sticky Sessions**: WebSocket affinity
- **Message Queue**: Decouple processing
- **Metrics**: Prometheus/Grafana monitoring

## Security Features

### Implemented
- Optional password protection per session
- Input validation on all messages
- Connection rate limiting (via reverse proxy)
- Public IP extraction from connection

### Recommended
- TLS/WSS (wss://) in production
- JWT-based authentication
- Rate limiting per IP
- DDoS protection (Cloudflare/GCP)
- TURN server fallback for restrictive NATs

## Deployment Architecture

```
┌─────────────────────────────────────────────┐
│          Google Cloud Run                    │
│                                              │
│  ┌────────────────────────────────────────┐ │
│  │  Container (signaling-server)          │ │
│  │  - Stateless                           │ │
│  │  - Auto-scaling                        │ │
│  │  - Health checks                       │ │
│  └────────────────────────────────────────┘ │
│                                              │
└─────────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────┐
│     Cloud Load Balancer (HTTPS)              │
│     - SSL Termination                        │
│     - WebSocket support                      │
└─────────────────────────────────────────────┘
```

## Performance Targets

- **Connection Handling**: 10K concurrent WebSockets
- **Message Latency**: <50ms p99
- **Session Lookup**: O(1) via sync.Map
- **Memory Usage**: ~10MB per 1K connections
- **CPU**: Minimal (<5% at 1K connections)

## Monitoring & Observability

### Metrics to Track
- Active connections count
- Active sessions count
- Message rate (per type)
- Error rate
- Connection duration
- Memory/CPU usage

### Logging
- Structured JSON logs
- Log levels: DEBUG, INFO, WARN, ERROR
- Context: connection_id, session_id, event_type

## NAT Traversal Strategy

1. **Public IP Discovery**: Extract from WebSocket connection
2. **Coordinate Exchange**: Share both local + public IPs
3. **UDP Hole Punching**: Peers attempt simultaneous UDP send
4. **Fallback**: TURN relay server (not implemented, recommended)

## Error Handling

- **Network Errors**: Graceful connection close, cleanup
- **Invalid Messages**: Return error response, keep connection
- **Session Conflicts**: Prevent duplicate session IDs
- **Timeout**: Auto-cleanup after 1 hour inactivity

## Testing Strategy

1. **Unit Tests**: Individual components (session, message parsing)
2. **Integration Tests**: Full message flow simulation
3. **Load Tests**: 1K+ concurrent connections (using `ws` tool)
4. **Manual Tests**: Test clients (provided)

## Future Enhancements

- [ ] Redis session storage
- [ ] JWT authentication
- [ ] Rate limiting middleware
- [ ] Prometheus metrics endpoint
- [ ] TURN server integration
- [ ] Admin API (list sessions, force disconnect)
- [ ] Session recording/analytics
- [ ] Multi-region deployment
