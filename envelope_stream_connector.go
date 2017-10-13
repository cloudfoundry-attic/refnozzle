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

type EnvelopeStreamConnector struct {
	addr    string
	tlsConf *tls.Config
}

func NewEnvelopeStreamConnector(
	addr string,
	c *tls.Config,
) *EnvelopeStreamConnector {
	return &EnvelopeStreamConnector{
		addr:    addr,
		tlsConf: c,
	}
}

type EnvelopeStream func() []*loggregator_v2.Envelope

func (c *EnvelopeStreamConnector) Stream(ctx context.Context, req *loggregator_v2.EgressBatchRequest) EnvelopeStream {
	return newStream(ctx, c.addr, req, c.tlsConf).recv
}

type stream struct {
	ctx    context.Context
	req    *loggregator_v2.EgressBatchRequest
	client loggregator_v2.EgressClient
	rx     loggregator_v2.Egress_BatchedReceiverClient
}

func newStream(ctx context.Context, addr string, req *loggregator_v2.EgressBatchRequest, c *tls.Config) *stream {
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

	return &stream{
		ctx:    ctx,
		req:    req,
		client: client,
	}
}

func (s *stream) recv() []*loggregator_v2.Envelope {
	for {
		ok := s.connect(s.ctx)
		if !ok {
			return nil
		}
		batch, err := s.rx.Recv()
		if err != nil {
			s.rx = nil
			continue
		}

		return batch.Batch
	}
}

func (c *stream) connect(ctx context.Context) bool {
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
