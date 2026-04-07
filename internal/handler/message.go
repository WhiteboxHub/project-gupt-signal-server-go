package handler

import (
	"encoding/json"
	"log"

	"github.com/sampath/signaling-server/internal/connection"
	"github.com/sampath/signaling-server/internal/models"
	"github.com/sampath/signaling-server/internal/nat"
	"github.com/sampath/signaling-server/internal/session"
)

// Handler processes incoming signaling messages
type Handler struct {
	connMgr    *connection.Manager
	sessionMgr *session.Manager
}

// NewHandler creates a new message handler
func NewHandler(connMgr *connection.Manager, sessionMgr *session.Manager) *Handler {
	return &Handler{
		connMgr:    connMgr,
		sessionMgr: sessionMgr,
	}
}

// HandleMessage routes messages to appropriate handlers
func (h *Handler) HandleMessage(connID string, remoteAddr string, data []byte) {
	var msg models.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("[Handler] Failed to parse message from %s: %v", connID, err)
		h.sendError(connID, "Invalid message format")
		return
	}

	log.Printf("[Handler] Received message type=%s session_id=%s from conn=%s", msg.Type, msg.SessionID, connID)

	switch msg.Type {
	case models.MsgTypeRegister:
		h.handleRegister(connID, remoteAddr, &msg)
	case models.MsgTypeConnect:
		h.handleConnect(connID, remoteAddr, &msg)
	case models.MsgTypeHeartbeat:
		h.handleHeartbeat(connID)
	default:
		log.Printf("[Handler] Unknown message type: %s", msg.Type)
		h.sendError(connID, "Unknown message type")
	}
}

// handleRegister processes host registration
func (h *Handler) handleRegister(connID, remoteAddr string, msg *models.Message) {
	// Validate required fields
	if msg.SessionID == "" {
		h.sendError(connID, "session_id is required")
		return
	}

	// Extract public IP from connection
	publicIP, publicPort := nat.EnrichPeerInfo(msg.LocalIP, msg.LocalPort, remoteAddr)

	// Create peer info
	hostPeer := &models.PeerInfo{
		LocalIP:    msg.LocalIP,
		LocalPort:  msg.LocalPort,
		PublicIP:   publicIP,
		PublicPort: publicPort,
	}

	// Create session
	err := h.sessionMgr.CreateSession(msg.SessionID, msg.Password, hostPeer, connID)
	if err != nil {
		log.Printf("[Handler] Failed to create session %s: %v", msg.SessionID, err)
		h.sendError(connID, err.Error())
		return
	}

	log.Printf("[Handler] Host registered session=%s conn=%s local=%s:%d public=%s:%d",
		msg.SessionID, connID, msg.LocalIP, msg.LocalPort, publicIP, publicPort)

	// Send success response
	response := models.SuccessResponse(models.MsgTypeRegistered, msg.SessionID)
	h.connMgr.SendMessage(connID, response)
}

// handleConnect processes client connection requests
func (h *Handler) handleConnect(connID, remoteAddr string, msg *models.Message) {
	// Validate required fields
	if msg.SessionID == "" {
		h.sendError(connID, "session_id is required")
		return
	}

	// Extract public IP from connection
	publicIP, publicPort := nat.EnrichPeerInfo(msg.LocalIP, msg.LocalPort, remoteAddr)

	// Create peer info
	clientPeer := &models.PeerInfo{
		LocalIP:    msg.LocalIP,
		LocalPort:  msg.LocalPort,
		PublicIP:   publicIP,
		PublicPort: publicPort,
	}

	// Join session
	sess, err := h.sessionMgr.JoinSession(msg.SessionID, msg.Password, clientPeer, connID)
	if err != nil {
		log.Printf("[Handler] Failed to join session %s: %v", msg.SessionID, err)
		h.sendError(connID, err.Error())
		return
	}

	log.Printf("[Handler] Client joined session=%s conn=%s local=%s:%d public=%s:%d",
		msg.SessionID, connID, msg.LocalIP, msg.LocalPort, publicIP, publicPort)

	// Exchange peer information
	// Send host info to client
	hostInfoMsg := models.PeerInfoResponse(sess.Host)
	h.connMgr.SendMessage(connID, hostInfoMsg)

	// Send client info to host
	clientInfoMsg := models.PeerInfoResponse(sess.Client)
	h.connMgr.SendMessage(sess.HostConnID, clientInfoMsg)

	log.Printf("[Handler] Peer exchange completed for session=%s", msg.SessionID)
}

// handleHeartbeat responds to heartbeat messages
func (h *Handler) handleHeartbeat(connID string) {
	// Simply acknowledge heartbeat
	response := &models.Message{
		Type: models.MsgTypeHeartbeat,
	}
	h.connMgr.SendMessage(connID, response)
}

// handleDisconnect cleans up when a connection is closed
func (h *Handler) HandleDisconnect(connID string) {
	log.Printf("[Handler] Connection disconnected: %s", connID)

	// Find and cleanup associated session
	sess := h.sessionMgr.GetSessionByConnID(connID)
	if sess != nil {
		log.Printf("[Handler] Cleaning up session %s due to disconnect", sess.ID)
		h.sessionMgr.DeleteSession(sess.ID)

		// Close the other peer's connection if exists
		if sess.HostConnID == connID && sess.ClientConnID != "" {
			h.connMgr.CloseConnection(sess.ClientConnID)
		} else if sess.ClientConnID == connID && sess.HostConnID != "" {
			h.connMgr.CloseConnection(sess.HostConnID)
		}
	}

	// Remove connection
	h.connMgr.RemoveConnection(connID)
}

// sendError sends an error message to a connection
func (h *Handler) sendError(connID, message string) {
	response := models.ErrorResponse(message)
	h.connMgr.SendMessage(connID, response)
}
