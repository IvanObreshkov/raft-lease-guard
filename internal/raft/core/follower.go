package core

// Follower is the passive state: it issues no requests of its own and only responds to leaders and candidates, as per
// Section 5.2 from the [Raft paper](https://raft.github.io/raft.pdf).
type Follower struct {
	// TicksUntilElection counts down on every Tick. At zero the follower starts an election, as per the Followers rules
	// in Figure 2 from the [Raft paper](https://raft.github.io/raft.pdf). It is refilled with a fresh random draw
	// whenever this server hears from the current leader or grants its vote.
	TicksUntilElection int
}

func (*Follower) isServerState() {}

// transition is the Followers box of Figure 2 from the [Raft paper](https://raft.github.io/raft.pdf).
func (st *Follower) transition(r *Raft, event Event) {
	switch event.(type) {
	case Tick:
		st.TicksUntilElection--
		if st.TicksUntilElection <= 0 {
			// TODO: r.becomeCandidate(), once candidate.go defines it.
		}
	case MessageReceived:
		// TODO(human): "Respond to RPCs from candidates and leaders". Apply the receiver rules for RequestVote and
		// AppendEntries from Figure 2 to this server's state, and refill st.TicksUntilElection with
		// r.drawElectionTicks() when the message is from the current leader or when this server grants its vote.
	}
}

// becomeFollower is "convert to follower" from the All Servers and Candidates rules in Figure 2 of the
// [Raft paper](https://raft.github.io/raft.pdf).
func (r *Raft) becomeFollower() {
	// TODO: set r.ServerState to a new &Follower{} with TicksUntilElection drawn from r.drawElectionTicks().
}
