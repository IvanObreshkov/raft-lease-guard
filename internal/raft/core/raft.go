package core

import (
	"time"

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

// RaftVolatileStateLeader is the non-persisted state of a Leader Raft server, it is reinitialized after an election
type RaftVolatileStateLeader struct {
	// NextIndex is, for each server, index of the next log entry to send to that server (initialized to leader last
	// log index + 1) as per Figure 2 from the [Raft paper](https://raft.github.io/raft.pdf)
	NextIndex map[ServerID]uint64
	// MatchIndex is, for each server, index of highest log entry known to be replicated on server (initialized to 0,
	// increases monotonically) as per Figure 2 from the [Raft paper](https://raft.github.io/raft.pdf)
	MatchIndex map[ServerID]uint64
}

// RaftState is the complete state of a Raft server: the persistent and volatile state that every server keeps, the volatile
// state kept only while it is the leader, and its current ServerState. See Figure 2 from the
// [Raft paper](https://raft.github.io/raft.pdf).
type RaftState struct {
	RaftPersistedState
	RaftVolatileState
	// LeaderState is the volatile leader state; it is non-nil only while State is Leader, and is reinitialized after
	// each election, as per Figure 2 from the [Raft paper](https://raft.github.io/raft.pdf).
	LeaderState *RaftVolatileStateLeader
	// State is the current state of the server: Follower, Candidate, or Leader, as per Section 5.1 from the
	// [Raft paper](https://raft.github.io/raft.pdf). A server starts as a Follower, as per Section 5.2.
	State ServerState
	// ElectionTimeout is the current election timeout for the server. It is randomly chosen when the server is created.
	// It should be used with a time.Timer, and the timer should be reset at the beginning of each new election and
	// when the server receives an AppendEntries RPC, as per Section 5.2 from the
	// [Raft paper](https://raft.github.io/raft.pdf). It only makes sense when Server is Follower or Candidate.
	ElectionTimeout time.Duration
}
// TODO: Add the Transition Func

type RaftServer struct {
	raftpb.UnimplementedRaftRPCServer
	State RaftState
	// TODO: Design the Transport layer interface and the main_loop of the server (one goroutine iterating over channels and calling Transition)
}