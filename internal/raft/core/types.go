package core

// ServerID is the ID of a server in the cluster
type ServerID int64

// ServerState is the state of a server at any given point: Follower, Candidate, or Leader, as per Section 5.1 from the
// [Raft paper](https://raft.github.io/raft.pdf).
type ServerState int

const (
	Follower ServerState = iota
	Leader
	Candidate
)
