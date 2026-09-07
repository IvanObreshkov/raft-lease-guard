package core

import (
	"context"
	"time"
)

// Server is the I/O shell around Raft: it owns the election timer, the transport, and the single goroutine that feeds
// messages to Raft.Transition and executes the actions it returns. Keeping the I/O here is what leaves Raft pure.
type Server struct {
	// Raft is the consensus core this server drives. Only the main loop calls into it, which is what makes it safe for
	// Transition to mutate the state without locking.
	Raft *Raft
	// ElectionTimeout is the current election timeout used to arm the election timer. It is drawn randomly from a fixed
	// interval and re-drawn at the start of each election (Section 5.2 of the [Raft paper](https://raft.github.io/raft.pdf))
	// so that split votes are rare. It lives on Server, not Raft, because only the timer reads it — the core
	// reacts to the timeout event, not its duration.
	ElectionTimeout time.Duration
	// Transport is how this server reaches its peers, kept as an interface so a simulated network can take the place of
	// the real one.
	Transport Transport
}

// mainLoop drives Raft from a single goroutine: it turns whatever happens into an Event, hands it to Transition and
// carries out the Effects that come back. Nothing else may call Transition, which is what makes mutating the state
// without a lock safe. It returns once ctx is cancelled.
func (s *Server) mainLoop(ctx context.Context) {
	electionTimer := time.NewTimer(s.ElectionTimeout)
	defer electionTimer.Stop()

}
