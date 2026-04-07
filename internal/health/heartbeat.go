package health

import (
	"log"
	"time"

	"github.com/sampath/signaling-server/internal/connection"
	"github.com/sampath/signaling-server/internal/session"
)

// Monitor handles periodic health checks and cleanup
type Monitor struct {
	connMgr    *connection.Manager
	sessionMgr *session.Manager
	interval   time.Duration
	stopCh     chan struct{}
}

// NewMonitor creates a new health monitor
func NewMonitor(connMgr *connection.Manager, sessionMgr *session.Manager, interval time.Duration) *Monitor {
	return &Monitor{
		connMgr:    connMgr,
		sessionMgr: sessionMgr,
		interval:   interval,
		stopCh:     make(chan struct{}),
	}
}

// Start begins the monitoring loop
func (m *Monitor) Start() {
	log.Printf("[HealthMonitor] Started with interval=%v", m.interval)

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanup()
		case <-m.stopCh:
			log.Println("[HealthMonitor] Stopped")
			return
		}
	}
}

// Stop stops the monitoring loop
func (m *Monitor) Stop() {
	close(m.stopCh)
}

// cleanup performs periodic cleanup tasks
func (m *Monitor) cleanup() {
	// Clean up expired sessions
	count := m.sessionMgr.CleanupExpiredSessions()
	if count > 0 {
		log.Printf("[HealthMonitor] Cleaned up %d expired sessions", count)
	}

	// Log current stats
	activeConns := m.connMgr.GetActiveConnections()
	activeSessions := m.sessionMgr.GetActiveSessions()

	log.Printf("[HealthMonitor] Stats: connections=%d sessions=%d", activeConns, activeSessions)
}
