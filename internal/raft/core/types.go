package core

import raftpb "github.com/IvanObreshkov/raft-lease-guard/internal/raft/proto"

// ServerID is the ID of a server in the cluster
type ServerID uint64

// ServerState is the state of a server at any given point: Follower, Candidate, or Leader, as per Section 5.1 from the
// [Raft paper](https://raft.github.io/raft.pdf). The set of variants is closed to this package, and the marker method
// has a pointer receiver, so only *Follower, *Candidate, and *Leader satisfy it.
//
//sumtype:decl
type ServerState interface {
	isServerState()
}

// Candidate is a server running an election for CurrentTerm, as per Section 5.2 from the
// [Raft paper](https://raft.github.io/raft.pdf).
type Candidate struct{}

func (*Candidate) isServerState() {}

// raftVolatileStateLeader is the non-persisted state of a Leader Raft server, it is reinitialized after an election
type raftVolatileStateLeader struct {
	// NextIndex is, for each server, index of the next log entry to send to that server (initialized to leader last
	// log index + 1) as per Figure 2 from the [Raft paper](https://raft.github.io/raft.pdf)
	NextIndex map[ServerID]uint64
	// MatchIndex is, for each server, index of highest log entry known to be replicated on server (initialized to 0,
	// increases monotonically) as per Figure 2 from the [Raft paper](https://raft.github.io/raft.pdf)
	MatchIndex map[ServerID]uint64
}

// Leader is the server that handles client requests and replicates the log, as per Section 5.1 from the
// [Raft paper](https://raft.github.io/raft.pdf).
type Leader struct {
	raftVolatileStateLeader
}

func (*Leader) isServerState() {}

// Event is one input to the state machine: a message that arrived from another server, or one tick of the clock. The
// set of variants is closed to this package, so a type switch in Transition covers every input there is.
//
// TODO: add the remaining variants, such as a command from a client.
//
//sumtype:decl
type Event interface {
	isEvent()
}

// MessageReceived is a RaftMessage that arrived from another server.
type MessageReceived struct {
	Message *raftpb.RaftMessage
}

func (MessageReceived) isEvent() {}

// Tick reports that one TickInterval of real time has passed. The core measures every timeout in Ticks and keeps no
// clock of its own.
type Tick struct{}

func (Tick) isEvent() {}
