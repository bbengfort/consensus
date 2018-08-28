/*
Package consensus implements a communication framework with gRPC streaming
between replicas that participate in a quorum. This framework is meant to
facilitate the implementation of consensus algorithms.
*/
package consensus

import (
	"fmt"
	"net"

	"github.com/bbengfort/consensus/pb"
	"google.golang.org/grpc"
)

// Replica objects implement the gRPC service and maintain configuration and
// state for responding to consensus remote proceedure calls.
type Replica struct {
	Name   string // unique name identifying the replica to peers
	IPAddr string // ip address or hostname the replica listens on
	Port   uint16 // the port the replica listens for requests on
}

// Listen connects the replica to all remote peers and runs the replica server
// to handle all incoming requests.
func (r *Replica) Listen() error {

	// Listen for requests from remote peers and clients on all addresses
	addr := fmt.Sprintf(":%d", r.Port)
	sock, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("could not listen on %s", addr)
	}
	defer sock.Close()
	fmt.Printf("listening for requests on %s\n", addr)

	// Initialize and run the gRPC server
	srv := grpc.NewServer()
	pb.RegisterConsensusServer(srv, r)

	return srv.Serve(sock)
}
