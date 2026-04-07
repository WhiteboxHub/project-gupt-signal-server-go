package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sampath/signaling-server/internal/connection"
	"github.com/sampath/signaling-server/internal/handler"
	"github.com/sampath/signaling-server/internal/health"
	"github.com/sampath/signaling-server/internal/session"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins (configure properly in production)
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Server represents the WebSocket signaling server
type Server struct {
	addr       string
	connMgr    *connection.Manager
	sessionMgr *session.Manager
	handler    *handler.Handler
	monitor    *health.Monitor
	httpServer *http.Server
}

// NewServer creates a new WebSocket server
func NewServer(addr string, sessionTTL time.Duration, monitorInterval time.Duration) *Server {
	connMgr := connection.NewManager()
	sessionMgr := session.NewManager(sessionTTL)
	msgHandler := handler.NewHandler(connMgr, sessionMgr)
	monitor := health.NewMonitor(connMgr, sessionMgr, monitorInterval)

	return &Server{
		addr:       addr,
		connMgr:    connMgr,
		sessionMgr: sessionMgr,
		handler:    msgHandler,
		monitor:    monitor,
	}
}

// Start starts the WebSocket server
func (s *Server) Start() error {
	// Start health monitor
	go s.monitor.Start()

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/health", s.handleHealth)

	s.httpServer = &http.Server{
		Addr:         s.addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("[Server] Starting on %s", s.addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("[Server] Shutting down gracefully...")
	s.monitor.Stop()

	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// handleRoot serves basic info
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Signaling Server\nConnect via WebSocket at /ws\n"))
}

// handleHealth returns server health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"healthy"}`))
}

// handleWebSocket handles WebSocket upgrade and connection
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[Server] WebSocket upgrade failed: %v", err)
		return
	}

	// Generate unique connection ID
	connID := uuid.New().String()
	remoteAddr := r.RemoteAddr

	log.Printf("[Server] New WebSocket connection: %s from %s", connID, remoteAddr)

	// Register connection
	s.connMgr.AddConnection(connID, conn)

	// Handle connection in a goroutine
	go s.handleConnection(connID, remoteAddr, conn)
}

// handleConnection manages a single WebSocket connection
func (s *Server) handleConnection(connID, remoteAddr string, conn *websocket.Conn) {
	defer func() {
		// Cleanup on disconnect
		s.handler.HandleDisconnect(connID)
		conn.Close()
	}()

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	// Setup pong handler to reset deadline
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Read messages in a loop
	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[Server] WebSocket error from %s: %v", connID, err)
			}
			break
		}

		// Only process text messages
		if messageType == websocket.TextMessage {
			s.handler.HandleMessage(connID, remoteAddr, data)
		}

		// Reset read deadline after successful read
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	}
}
