package models

import "time"

// MessageType represents the type of signaling message
type MessageType string

const (
	MsgTypeRegister   MessageType = "register"
	MsgTypeConnect    MessageType = "connect"
	MsgTypePeerInfo   MessageType = "peer_info"
	MsgTypeHeartbeat  MessageType = "heartbeat"
	MsgTypeError      MessageType = "error"
	MsgTypeRegistered MessageType = "registered"
	MsgTypeConnected  MessageType = "connected"
)

// Message represents a generic signaling message
type Message struct {
	Type      MessageType `json:"type"`
	SessionID string      `json:"session_id,omitempty"`
	Password  string      `json:"password,omitempty"`
	LocalIP   string      `json:"local_ip,omitempty"`
	LocalPort int         `json:"local_port,omitempty"`
	Peer      *PeerInfo   `json:"peer,omitempty"`
	Message   string      `json:"message,omitempty"`
}

// PeerInfo contains connection details for a peer
type PeerInfo struct {
	LocalIP    string `json:"local_ip"`
	LocalPort  int    `json:"local_port"`
	PublicIP   string `json:"public_ip"`
	PublicPort int    `json:"public_port"`
}

// SessionStatus represents the current state of a session
type SessionStatus string

const (
	SessionStatusWaiting  SessionStatus = "waiting"  // Host registered, waiting for client
	SessionStatusActive   SessionStatus = "active"   // Both peers connected
	SessionStatusExpired  SessionStatus = "expired"  // Session timed out
	SessionStatusClosed   SessionStatus = "closed"   // Session explicitly closed
)

// Session represents a signaling session between two peers
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

// ErrorResponse creates an error message
func ErrorResponse(msg string) *Message {
	return &Message{
		Type:    MsgTypeError,
		Message: msg,
	}
}

// SuccessResponse creates a success registration message
func SuccessResponse(msgType MessageType, sessionID string) *Message {
	return &Message{
		Type:      msgType,
		SessionID: sessionID,
	}
}

// PeerInfoResponse creates a peer info message
func PeerInfoResponse(peer *PeerInfo) *Message {
	return &Message{
		Type: MsgTypePeerInfo,
		Peer: peer,
	}
}
