package core

import (
	"context"

	raftpb "github.com/IvanObreshkov/raft-lease-guard/internal/raft/proto"
)

// Receiver hands over the messages that arrived from other servers.
type Receiver interface {
	// Inbound is selected on by the main loop, so it must not block while a message is waiting.
	Inbound() <-chan *raftpb.RaftMessage
}

// Sender delivers a message to the server named in its To field.
type Sender interface {
	Dispatch(ctx context.Context, msg *raftpb.RaftMessage) error
}

type Transport interface {
	Receiver
	Sender
}
