// Package core is the Raft state machine without any I/O: it consumes Events and mutates its state in place.
package core

import (
	"errors"
	"fmt"

	raftpb "github.com/IvanObreshkov/raft-lease-guard/internal/raft/proto"
)

// RaftPersistedState is the state of a Raft server persisted on stable storage before responding to RPCs
// See: Figure 2 from the [Raft paper](https://raft.github.io/raft.pdf)
type RaftPersistedState struct {
	// CurrentTerm is the latest term the server has seen, initialized to 0 on first boot of the cluster and increasing
	// monotonically, as per Figure 2 and Section 5.1 from the [Raft paper](https://raft.github.io/raft.pdf). It serves
	// as a [logical clock](https://dl.acm.org/doi/pdf/10.1145/359545.359563) that lets servers detect obsolete
	// information, such as a stale leader.
	CurrentTerm uint64
	// VotedFor is the candidateId (ServerID) that received this server's vote in the current term, or nil if it has not
	// voted, as per Figure 2 from the [Raft paper](https://raft.github.io/raft.pdf). Votes are per term, so it is reset
	// to nil whenever CurrentTerm changes.
	VotedFor *ServerID
	// Log holds the log entries; each entry contains a command for the replicated state machine and the term when the
	// entry was received by the leader. The first index is 1, as per Figure 2 and Section 5.3 from the
	// [Raft paper](https://raft.github.io/raft.pdf).
	Log []*raftpb.LogEntry
}

// RaftVolatileState is the non-persisted state of a Raft server, regardless of its current ServerState
type RaftVolatileState struct {
	// CommitIndex is the index of highest log entry known to be committed (initialized to 0, increases monotonically)
	// as per Figure 2 from the [Raft paper](https://raft.github.io/raft.pdf)
	CommitIndex uint64
	// LastApplied is the index of highest log entry applied to state machine (initialized to 0, increases monotonically)
	// as per Figure 2 from the [Raft paper](https://raft.github.io/raft.pdf)
	LastApplied uint64
}

// RaftState is the complete state of a Raft server: the persistent and volatile state that every server keeps, and its
// current ServerState, which carries whatever is specific to that state. See Figure 2 from the
// [Raft paper](https://raft.github.io/raft.pdf).
type RaftState struct {
	RaftPersistedState
	RaftVolatileState
	// ServerState is the current state of the server: *Follower, *Candidate, or *Leader, as per Section 5.1 from the
	// [Raft paper](https://raft.github.io/raft.pdf). It is never nil: a server starts as a Follower, as per Section 5.2.
	ServerState ServerState
}

// ElectionTicksDraw returns how many Ticks a new Follower or Candidate waits before starting an election. The result
// must be greater than 0. The Server supplies a random draw from [electionTicks, 2*electionTicks), as per Section 5.2
// from the [Raft paper](https://raft.github.io/raft.pdf).
type ElectionTicksDraw func() int

// Raft is the consensus core. It performs no I/O and reads no clock of its own: Transition mutates this state in place
// in response to the events the Server feeds it.
type Raft struct {
	RaftState
	// drawElectionTicks is the only source of election timeouts. The core holds no random source of its own.
	drawElectionTicks ElectionTicksDraw
	// heartbeatTicks is how many Ticks a Leader lets pass between heartbeats. Always greater than 0.
	heartbeatTicks int
}

// NewRaft returns a server that starts as a Follower, as per Section 5.2 from the
// [Raft paper](https://raft.github.io/raft.pdf), with its first election timeout already drawn. It rejects a nil draw
// and a non-positive heartbeatTicks here, because either would otherwise surface as a leader or follower that never
// times out.
func NewRaft(drawElectionTicks ElectionTicksDraw, heartbeatTicks int) (*Raft, error) {
	if drawElectionTicks == nil {
		return nil, errors.New("drawElectionTicks must not be nil")
	}
	if heartbeatTicks <= 0 {
		return nil, fmt.Errorf("heartbeatTicks must be greater than 0, got %d", heartbeatTicks)
	}
	r := &Raft{drawElectionTicks: drawElectionTicks, heartbeatTicks: heartbeatTicks}
	r.ServerState = &Follower{TicksUntilElection: r.drawElectionTicks()}
	return r, nil
}

// Transition applies event to this server, as per the Rules for Servers in Figure 2 from the
// [Raft paper](https://raft.github.io/raft.pdf). It mutates the state in place and performs no I/O, so the only caller
// may be the single goroutine that owns this Raft.
func (r *Raft) Transition(event Event) {
	// TODO: the All Servers rules from Figure 2. A message whose term is greater than CurrentTerm makes this server
	// adopt that term and become a Follower before the dispatch below, so the message is then handled as a Follower
	// whatever this server was a moment ago.

	switch st := r.ServerState.(type) {
	case *Follower:
		st.transition(r, event)
	case *Candidate:
		// TODO: st.transition(r, event), once candidate.go defines it.
	case *Leader:
		// TODO: st.transition(r, event), once leader.go defines it.
	}
}
