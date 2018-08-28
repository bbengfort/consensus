# Consensus Communication

**Consensus communication stubs with gRPC streaming.**

## RPC and Communication

Communication between replicas involves messages sent as [protocol buffers](https://developers.google.com/protocol-buffers/docs/proto3) over [gRPC](https://grpc.io/docs/quickstart/go.html) HTTP connections. The `github.com/bbengfort/consensus/pb` package implements several message types and a service in `.proto` files.

> **NOTE**: only edit the `pb/*.proto` files, the `pb/*.pb.go` are generated using the `make protobuf` command, which will overwrite any changes made. If you need to add Go to the `pb` package, please do so in `pb/*.go` files.

RPCs between replicas and clients is defined by two gRPC services:

- **Propose**: a unary RPC that allows clients to propose commands and values to any replica to be committed by the quorum.
- **Dispatch**: a bidirectional streaming RPC that allows replicas to send generic messages to each other.

In order to enable communication, a replica/server must implement the `pb.ConsensusServer` interface, defined as follows:

```go
type ConsensusServer interface {
	Propose(ctx context.Context, req *ProposeRequest) (*ProposeReply, error)
	Dispatch(stream Consensus_DispatchServer) error
}
```

The `Dispatch` method will receive `*pb.PeerRequest` messages, a wrapper message defined in `pb/envelope.proto`. The wrapper message specifies the type of message, and can contain _one_ wrapped message. To read a message from the stream, you would check it's type, then use the appropriate `Get` method:

```go
import (
	"fmt"

	"github.com/bbengfort/consensus/pb"
)

func ExampleParseBeacon(req *pb.PeerRequest) (*pb.BeaconRequest, error) {
	if req.GetType() == pb.Type_BEACON {
		return req.GetBeacon(), nil
	}

	return nil, fmt.Errorf("message with type %d is not a beacon request", req.GetType())
}
```

A `Dispatch` method therefore simply handle each incoming message with its own method based on it's type then send a reply, which may look like:

```go
func (r *Replica) Dispatch(stream pb.Consensus_DispatchServer) (err error) {
	// Receive each message from the stream, handle it, then send the reply.
	for {
		var req *pb.PeerRequest
		if req, err = stream.Recv(); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		// Handle the specific event type and get a reply
		var rep *pb.PeerReply
		switch req.GetType() {
		case pb.Type_BEACON:
			rep, err = r.Beacon(req)
		case pb.Type_PREPARE:
			rep, err = r.Prepare(req)
		}

		// Exit if there is an error handling the message
		if err != nil {
			return err
		}

		// Send the reply on the stream and handle the next message
		if err = stream.Send(rep); err != nil {
			return err
		}

	}
}
```

To create a beacon reply to send on the stream, you have to assign the message to the correct type with a `pb.PeerReply_[Type]` wrapper as follows (this is also how you might create `pb.PeerRequest` messages as well):

```go
import (
	"os"
	"time"

	"github.com/bbengfort/consensus/pb"
)

func ExampleCreateBeaconReply() *pb.PeerReply {
	hostname, _ := os.Hostname()

	req := &pb.PeerReply{
		Type:   pb.Type_BEACON,
		Sender: hostname,
		Message: &pb.PeerReply_Beacon{
			Beacon: &pb.BeaconReply{
				Timestamp: time.Now().Format(time.RFC3339Nano),
			},
		},
	}

	return req
}
```

> **NOTE**: the message types defined in `pb/*.proto` are based on the protocol buffers defined in the [original implementation of ePaxos](https://github.com/efficient/epaxos). They are meant to be an example only, and can be modified and recompiled as needed.

Adding a new message type to the `Dispatch` streaming RPC is fairly straight forward; create the new message request and reply types in a `pb/.proto` file, then import it in `pb/envelope.proto`. In that same file, add the type to the `Type` enumeration, then the appropriate fields to the `PeerRequest` and `PeerReply` message `oneof` field. Rebuild the protocol buffers as follows:

```
$ make protobuf
```

Note that to run this command, [Protocol Buffers v3](https://grpc.io/docs/quickstart/go.html#install-protocol-buffers-v3) must be installed on your system (this is not a go dependency).
