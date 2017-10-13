package refnozzle

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Connector struct {
	conn   *grpc.ClientConn
	client loggregator_v2.EgressClient
	rx     loggregator_v2.Egress_BatchedReceiverClient
	req    *loggregator_v2.EgressBatchRequest
}

func NewConnector(
	req *loggregator_v2.EgressBatchRequest,
	addr string,
	c *tls.Config,
) *Connector {
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(credentials.NewTLS(c)),
	)
	if err != nil {
		// This error occurs on invalid configuration. And more notably,
		// it does NOT occur if the server is not up.
		log.Panicf("Invalid gRPC dial configuration: %s", err)
	}

	client := loggregator_v2.NewEgressClient(conn)

	return &Connector{
		req:    req,
		conn:   conn,
		client: client,
	}
}

func (c *Connector) Receive() []*loggregator_v2.Envelope {
	for {
		ok := c.connect(context.TODO())
		if !ok {
			return nil
		}
		batch, err := c.rx.Recv()
		if err != nil {
			c.rx = nil
			continue
		}

		return batch.Batch
	}
}

func (c *Connector) connect(ctx context.Context) bool {
	for {
		select {
		case <-ctx.Done():
			return false
		default:
			if c.rx != nil {
				return true
			}

			var err error
			c.rx, err = c.client.BatchedReceiver(
				ctx,
				c.req,
			)

			if err != nil {
				log.Println("Error connecting to Logs Provider: %s", err)
				time.Sleep(50 * time.Millisecond)
				continue
			}

			return true
		}
	}
}
