package session

import (
	"errors"
	"sync"
	"time"

	"github.com/sampath/signaling-server/internal/models"
)

var (
	ErrSessionNotFound      = errors.New("session not found")
	ErrSessionAlreadyExists = errors.New("session already exists")
	ErrInvalidPassword      = errors.New("invalid password")
	ErrSessionFull          = errors.New("session already has a client")
)

// Manager handles session lifecycle and storage
type Manager struct {
	sessions sync.Map // map[string]*models.Session
	ttl      time.Duration
}

// NewManager creates a new session manager
func NewManager(ttl time.Duration) *Manager {
	return &Manager{
		ttl: ttl,
	}
}

// CreateSession creates a new session with a host
func (m *Manager) CreateSession(sessionID, password string, host *models.PeerInfo, hostConnID string) error {
	// Check if session already exists
	if _, exists := m.sessions.Load(sessionID); exists {
		return ErrSessionAlreadyExists
	}

	session := &models.Session{
		ID:         sessionID,
		Password:   password,
		Host:       host,
		HostConnID: hostConnID,
		CreatedAt:  time.Now(),
		Status:     models.SessionStatusWaiting,
	}

	m.sessions.Store(sessionID, session)
	return nil
}

// JoinSession adds a client to an existing session
func (m *Manager) JoinSession(sessionID, password string, client *models.PeerInfo, clientConnID string) (*models.Session, error) {
	value, exists := m.sessions.Load(sessionID)
	if !exists {
		return nil, ErrSessionNotFound
	}

	session := value.(*models.Session)

	// Validate password if set
	if session.Password != "" && session.Password != password {
		return nil, ErrInvalidPassword
	}

	// Check if session already has a client
	if session.Client != nil {
		return nil, ErrSessionFull
	}

	// Update session with client info
	session.Client = client
	session.ClientConnID = clientConnID
	session.Status = models.SessionStatusActive

	m.sessions.Store(sessionID, session)
	return session, nil
}

// GetSession retrieves a session by ID
func (m *Manager) GetSession(sessionID string) (*models.Session, error) {
	value, exists := m.sessions.Load(sessionID)
	if !exists {
		return nil, ErrSessionNotFound
	}
	return value.(*models.Session), nil
}

// DeleteSession removes a session
func (m *Manager) DeleteSession(sessionID string) {
	m.sessions.Delete(sessionID)
}

// CleanupExpiredSessions removes sessions older than TTL
func (m *Manager) CleanupExpiredSessions() int {
	count := 0
	now := time.Now()

	m.sessions.Range(func(key, value interface{}) bool {
		session := value.(*models.Session)
		if now.Sub(session.CreatedAt) > m.ttl {
			m.sessions.Delete(key)
			count++
		}
		return true
	})

	return count
}

// GetSessionByConnID finds a session by connection ID (host or client)
func (m *Manager) GetSessionByConnID(connID string) *models.Session {
	var foundSession *models.Session

	m.sessions.Range(func(key, value interface{}) bool {
		session := value.(*models.Session)
		if session.HostConnID == connID || session.ClientConnID == connID {
			foundSession = session
			return false // Stop iteration
		}
		return true
	})

	return foundSession
}

// GetActiveSessions returns count of active sessions
func (m *Manager) GetActiveSessions() int {
	count := 0
	m.sessions.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}
