package transport

import (
	"context"
	"errors"
	"fmt"

	"github.com/IvanObreshkov/raft-lease-guard/internal/raft/core"
	raftpb "github.com/IvanObreshkov/raft-lease-guard/internal/raft/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// gRPCInboundTransport receives by serving peers' Send calls: nothing is called to receive, messages arrive on Inbound.

type gRPCInboundTransport struct {
	// Required to implement raftpb.RaftRPCServer.
	raftpb.UnimplementedRaftRPCServer
	grpcServer *grpc.Server
	// inbound is read only by the main loop, keeping one goroutine the sole caller of the state machine.
	inbound chan *raftpb.RaftMessage
}

// Send forwards a peer's message to the main loop, blocking until the loop takes it so a busy node pushes back rather
// than drops. The Empty only acknowledges delivery; a Raft response comes later as a separate Send to the sender.
func (t *gRPCInboundTransport) Send(ctx context.Context, msg *raftpb.RaftMessage) (*emptypb.Empty, error) {
	select {
	case t.inbound <- msg:
		return &emptypb.Empty{}, nil
	case <-ctx.Done():
		return nil, status.FromContextError(ctx.Err()).Err()
	}
}

// Inbound is receive-only so that only the Send handler above can put messages on it.
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

type gRPCOutboundTransport struct {
	// selfID lets Dispatch skip a message meant for this server instead of sending it over the network.
	selfID core.ServerID
	// peers holds the other servers; each connection stays idle until the first message to it.
	peers map[core.ServerID]*grpc.ClientConn
}

// NewGRPCOutboundTransport creates a lazy connection to each of the other servers, so a bad address only surfaces on
// the first message to it. An address goes to grpc.NewClient and is resolved by its dns resolver, so "10.0.0.2:8080"
// pins a host and a plain "raft-2.raft-headless:8080" follows a changing IP.
func NewGRPCOutboundTransport(selfID core.ServerID, addresses map[core.ServerID]string) (*gRPCOutboundTransport, error) {
	peers := make(map[core.ServerID]*grpc.ClientConn, len(addresses))
	for id, address := range addresses {
		conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, fmt.Errorf("create client for server %d at %s: %w", id, address, err)
		}

		peers[id] = conn
	}

	return &gRPCOutboundTransport{selfID: selfID, peers: peers}, nil
}

// Dispatch sends msg to the server in its To field and waits, so it stalls the main loop for as long as the network
// takes. A message meant for this server is skipped rather than sent, so a caller can loop over every server in the
// cluster without filtering itself out. It errors when msg.To has no connection, which is a caller bug, and on a failed
// delivery, which is safe to ignore, as Raft resends what a peer missed.
func (t *gRPCOutboundTransport) Dispatch(ctx context.Context, msg *raftpb.RaftMessage) error {
	to := core.ServerID(msg.GetTo())
	if to == t.selfID {
		return nil
	}

	conn, ok := t.peers[to]
	if !ok {
		return fmt.Errorf("no peer with ID %d", to)
	}

	_, err := raftpb.NewRaftRPCClient(conn).Send(ctx, msg)

	return err
}

// Close closes every peer connection, returning all failures rather than stopping at the first.
func (t *gRPCOutboundTransport) Close() error {
	var closeErrors []error
	for id, conn := range t.peers {
		if err := conn.Close(); err != nil {
			closeErrors = append(closeErrors, fmt.Errorf("close connection to server %d: %w", id, err))
		}
	}

	return errors.Join(closeErrors...)
}

// GRPCTransport puts both halves behind one value, so a Server can be handed a single core.Transport instead of a
// sending half and a receiving half.
type GRPCTransport struct {
	*gRPCInboundTransport
	*gRPCOutboundTransport
}

// NewGRPCTransport builds both halves of the transport for the server identified by selfID. See
// NewGRPCInboundTransport and NewGRPCOutboundTransport for what the arguments mean. The gRPC server still has to be
// served on a listener before any peer can reach this one.
func NewGRPCTransport(
	selfID core.ServerID,
	inboundChannelSize int,
	addresses map[core.ServerID]string,
) (*GRPCTransport, error) {
	outbound, err := NewGRPCOutboundTransport(selfID, addresses)
	if err != nil {
		return nil, err
	}

	return &GRPCTransport{
		gRPCInboundTransport:  NewGRPCInboundTransport(inboundChannelSize),
		gRPCOutboundTransport: outbound,
	}, nil
}

// Close stops serving peers and closes every connection to them. It is spelled out rather than promoted, as otherwise
// only the outbound half would be closed and the gRPC server would keep running.
func (t *GRPCTransport) Close() error {
	t.grpcServer.Stop()

	return t.gRPCOutboundTransport.Close()
}
