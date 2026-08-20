package transport

import (
	"context"

	raftpb "github.com/IvanObreshkov/raft-lease-guard/internal/raft/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// gRPCInboundTransport receives RaftMessages from peers and hands them to the single goroutine that drives the Raft
// state machine. A node does not "receive" by calling anything: it receives by serving its peers' Send calls, and every
// message a peer delivers is forwarded on the Inbound channel.
type gRPCInboundTransport struct {
	// This makes the Server struct impl the proto.RaftServiceServer interface
	raftpb.UnimplementedRaftRPCServer
	// The underlying gRPC server used for receiving RPC messages
	grpcServer *grpc.Server
	// inbound carries every RaftMessage received from a peer to the main loop. Only the main loop reads from it, which
	// is what keeps a single goroutine as the sole caller of the Raft state machine.
	inbound chan *raftpb.RaftMessage
}

// Send handles a RaftMessage delivered by a peer by forwarding it to the main loop. It blocks until the loop accepts
// the message, so that a node under load pushes back on its peers rather than dropping messages. The returned Empty
// only acknowledges delivery: the Raft response, if any, is sent later as a separate Send call from this node back to
// the sender.
func (t *gRPCInboundTransport) Send(ctx context.Context, msg *raftpb.RaftMessage) (*emptypb.Empty, error) {
	select {
	case t.inbound <- msg:
		return &emptypb.Empty{}, nil
	case <-ctx.Done():
		return nil, status.FromContextError(ctx.Err()).Err()
	}
}

// Inbound returns the channel of RaftMessages received from peers, for the main loop to select on alongside its other
// event sources. It is receive-only, as the main loop is the only reader and the Send handler is the only sender.
func (t *gRPCInboundTransport) Inbound() <-chan *raftpb.RaftMessage {
	return t.inbound
}

func NewGRPCInboundTransport(inboundChannelSize int) *gRPCInboundTransport {
	transport := &gRPCInboundTransport{
		grpcServer: grpc.NewServer(),
		inbound:    make(chan *raftpb.RaftMessage, inboundChannelSize),
	}
	raftpb.RegisterRaftRPCServer(transport.grpcServer, transport)

	return transport
}
