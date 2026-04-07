package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type      string    `json:"type"`
	SessionID string    `json:"session_id,omitempty"`
	Password  string    `json:"password,omitempty"`
	LocalIP   string    `json:"local_ip,omitempty"`
	LocalPort int       `json:"local_port,omitempty"`
	Peer      *PeerInfo `json:"peer,omitempty"`
	Message   string    `json:"message,omitempty"`
}

type PeerInfo struct {
	LocalIP    string `json:"local_ip"`
	LocalPort  int    `json:"local_port"`
	PublicIP   string `json:"public_ip"`
	PublicPort int    `json:"public_port"`
}

func main() {
	serverURL := flag.String("server", "ws://localhost:8080/ws", "WebSocket server URL")
	sessionID := flag.String("session", "TEST123", "Session ID")
	password := flag.String("password", "", "Optional password")
	localIP := flag.String("local-ip", "192.168.1.200", "Local IP address")
	localPort := flag.Int("local-port", 9001, "Local port")
	flag.Parse()

	log.SetFlags(log.Ltime)
	log.Printf("=== CLIENT TEST CLIENT ===")
	log.Printf("Server: %s", *serverURL)
	log.Printf("Session: %s", *sessionID)
	log.Println("==========================")

	// Connect to WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(*serverURL, nil)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	log.Println("Connected to server")

	// Setup interrupt handler
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Channel for reading messages
	done := make(chan struct{})

	// Read messages from server
	go func() {
		defer close(done)
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Read error: %v", err)
				return
			}

			var msg Message
			if err := json.Unmarshal(data, &msg); err != nil {
				log.Printf("Parse error: %v", err)
				continue
			}

			log.Printf("Received: type=%s", msg.Type)

			if msg.Type == "peer_info" && msg.Peer != nil {
				log.Printf("=== PEER INFORMATION ===")
				log.Printf("Host Local IP:  %s:%d", msg.Peer.LocalIP, msg.Peer.LocalPort)
				log.Printf("Host Public IP: %s:%d", msg.Peer.PublicIP, msg.Peer.PublicPort)
				log.Println("========================")
				log.Println("You can now establish P2P connection using these details!")
			}

			if msg.Type == "error" {
				log.Printf("ERROR: %s", msg.Message)
			}
		}
	}()

	// Send connect message
	connectMsg := Message{
		Type:      "connect",
		SessionID: *sessionID,
		Password:  *password,
		LocalIP:   *localIP,
		LocalPort: *localPort,
	}

	data, _ := json.Marshal(connectMsg)
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Fatalf("Failed to send connect message: %v", err)
	}

	log.Printf("Sent connect message (session=%s)", *sessionID)

	// Send heartbeat every 30 seconds
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			log.Println("Connection closed")
			return
		case <-interrupt:
			log.Println("Interrupt received, closing connection...")
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		case <-ticker.C:
			heartbeat := Message{Type: "heartbeat"}
			data, _ := json.Marshal(heartbeat)
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("Heartbeat failed: %v", err)
				return
			}
		}
	}
}
