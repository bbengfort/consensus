package consensus

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/bbengfort/consensus/pb"
)

// Propose is a unary RPC that allows clients to propose commands to the quorum.
func (r *Replica) Propose(ctx context.Context, req *pb.ProposeRequest) (*pb.ProposeReply, error) {
	// Record the client request (optional)
	r.Metrics.Request(req.Identity)

	reply := &pb.ProposeReply{
		Success: false,
		Error:   "RPC not implemented yet",
	}

	// Record the client response (optional)
	r.Metrics.Complete(false)
	return reply, nil
}

// Beacon messages are sent to confirm liveness of peers.
func (r *Replica) Beacon(req *pb.BeaconRequest) (*pb.PeerReply, error) {
	reply := &pb.PeerReply{
		Type:   pb.Type_BEACON,
		Sender: r.Name,
		Message: &pb.PeerReply_Beacon{
			Beacon: &pb.BeaconReply{
				Timestamp: time.Now().Format(time.RFC3339Nano),
			},
		},
	}

	return reply, nil
}

// Dispatch is called when a remote peer connects a bidirectional stream to
// the local server. The stream receives *pb.PeerRequest objects, which wrap
// multiple message types defined by req.GetType() and accessed using the
// appropriate req.Get method. The stream sends *pb.PeerReply messages back
// to the remote peer.
//
// The implementation of this method is to determine the type, then pass the
// type to an appropriate handler method, which is responsible for creating
// the *pb.PeerReply. If the handler method returns an error, it is considered
// fatal, and the stream is closed. The method also0 ensures that one reply is
// sent for every request, in order of the replies received -- this is usually
// an important invariant for correctness in distributed consensus. However,
// this also means that all responses to the remote peer will be blocked while
// the current message is being handled.
func (r *Replica) Dispatch(stream pb.Consensus_DispatchServer) (err error) {
	// Receive each message from the stream, handle it, then send the reply.
	for {
		var req *pb.PeerRequest
		if req, err = stream.Recv(); err != nil {
			if err == io.EOF {
				// the stream was closed by the remote peer, return no error
				return nil
			}
			return err
		}

		// Handle the specific message as needed and return the reply.
		// TODO: add other handlers as needed.
		var rep *pb.PeerReply
		switch req.GetType() {
		case pb.Type_BEACON:
			rep, err = r.Beacon(req.GetBeacon())
		default:
			err = fmt.Errorf("message type '%s' handler is not implemented", req.GetType())
		}

		// Close the connection to the remote peer if there is a fatal error.
		if err != nil {
			return err
		}

		// Send the reply on the stream and handle the next message
		if err = stream.Send(rep); err != nil {
			return err
		}

	}
}
