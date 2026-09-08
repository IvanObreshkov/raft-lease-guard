// Package raft wraps core in the I/O it needs: a clock, a transport, and the goroutine that drives them.
package raft

import (
	"context"
	"time"

	"github.com/IvanObreshkov/raft-lease-guard/internal/raft/core"
)

// Server is the I/O shell around core.Raft: it owns the clock, the transport, and the single goroutine that feeds events
// to Transition. Keeping the I/O here is what leaves the core pure.
type Server struct {
	// Raft is the consensus core this server drives. Only the main loop calls into it, which is what makes it safe for
	// Transition to mutate the state without locking.
	Raft *core.Raft
	// TickInterval is the real-time length of one core.Tick. Every timeout in the core is a count of Ticks, so this is
	// the one place that decides how long ElectionTicks and HeartbeatTicks really take.
	TickInterval time.Duration
	Transport    Transport
}

// mainLoop drives the core from a single goroutine: it turns whatever happens into a core.Event and hands it to
// Transition. Nothing else may call Transition, which is what makes mutating the state without a lock safe. It returns
// once ctx is cancelled.
func (s *Server) mainLoop(ctx context.Context) {
	ticker := time.NewTicker(s.TickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.Raft.Transition(core.Tick{})
		case message := <-s.Transport.Inbound():
			s.Raft.Transition(core.MessageReceived{Message: message})
		case <-ctx.Done():
			return
		}
	}
}
