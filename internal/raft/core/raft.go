package core

import (
	raftpb "github.com/IvanObreshkov/raft-lease-guard/internal/raft/proto"
)

// RaftPersistedState is the state of a Raft server persisted on stable storage before responding to RPCs
// See: Figure 2 from the [Raft paper](https://raft.github.io/raft.pdf)
type RaftPersistedState struct {
	// CurrentTerm is the latest term the server has seen, initialized to 0 on first boot of the cluster and increasing
	// monotonically, as per Figure 2 and Section 5.1 from the [Raft paper](https://raft.github.io/raft.pdf). It serves
	// as a [logical clock](https://dl.acm.org/doi/pdf/10.1145/359545.359563) that lets servers detect obsolete
	// information, such as a stale leader.
	CurrentTerm int64
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

// RaftState is the complete state of a Raft server: the persistent and volatile state that every server keeps, the volatile
// state kept only while it is the leader, and its current ServerState. See Figure 2 from the
// [Raft paper](https://raft.github.io/raft.pdf).
type RaftState struct {
	RaftPersistedState
	RaftVolatileState
	// State is the current state of the server: Follower, Candidate, or Leader, as per Section 5.1 from the
	// [Raft paper](https://raft.github.io/raft.pdf). A server starts as a Follower, as per Section 5.2.
	State ServerState
}

// Raft is the consensus core, It performs no I/O of its own; Transition mutates this state and returns the
// actions for a Server to carry out.
type Raft struct {
	RaftState
	// ID is this server's own ID. A transition needs it to vote for itself when starting an election, as per Section 5.2
	// from the [Raft paper](https://raft.github.io/raft.pdf), and to stamp `from` on every outgoing RaftMessage.
	ID ServerID
	// Servers are all the servers in the cluster, including this one. Counting a majority is therefore
	// len(Servers)/2 + 1, with no adjustment for self, and the leader's own log counts towards a commit like any
	// follower's. Only the loops that "send RPCs to all other servers" (Section 5.2 from the
	// [Raft paper](https://raft.github.io/raft.pdf)) skip ID. It is a slice and not a map so the iteration order is
	// fixed, which keeps the actions a transition returns deterministic.
	Servers []ServerID
}

// Transition applies event to this server and reports what the Server must do as a result, as per the Rules for Servers
// in Figure 2 of the [Raft paper](https://raft.github.io/raft.pdf). It mutates the state in place and performs no I/O,
// so the only caller may be the single goroutine that owns this Raft.
func (r *Raft) Transition(event Event) Effects {
	// TODO: implement the Rules for Servers.
	return Effects{}
}
