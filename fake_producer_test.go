package refnozzle_test

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"sync"
	"time"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
)

type fakeEventProducer struct {
	server *grpc.Server
	addr   string

	mu                  sync.Mutex
	connectionAttempts_ int
	actualReq_          *loggregator_v2.EgressBatchRequest
}

func newFakeEventProducer() (*fakeEventProducer, error) {
	f := &fakeEventProducer{}

	return f, nil
}

func (f *fakeEventProducer) Receiver(
	*loggregator_v2.EgressRequest,
	loggregator_v2.Egress_ReceiverServer,
) error {

	return grpc.Errorf(codes.Unimplemented, "use BatchedReceiver instead")
}

func (f *fakeEventProducer) BatchedReceiver(
	req *loggregator_v2.EgressBatchRequest,
	srv loggregator_v2.Egress_BatchedReceiverServer,
) error {
	f.mu.Lock()
	f.connectionAttempts_++
	f.actualReq_ = req
	f.mu.Unlock()
	var i int
	for range time.Tick(10 * time.Millisecond) {
		srv.Send(&loggregator_v2.EnvelopeBatch{
			Batch: []*loggregator_v2.Envelope{
				{
					SourceId: fmt.Sprintf("envelope-%d", i),
					Message: &loggregator_v2.Envelope_Event{
						Event: &loggregator_v2.Event{
							Title: "event-name",
							Body:  "event-body",
						},
					},
				},
			},
		})
		i++
	}
	return nil
}

func (f *fakeEventProducer) start() {
	if f.addr == "" {
		f.addr = "localhost:0"
	}
	lis, err := net.Listen("tcp", f.addr)
	if err != nil {
		panic(err)
	}
	f.addr = lis.Addr().String()
	c, err := newServerMutualTLSConfig()
	if err != nil {
		panic(err)
	}
	opt := grpc.Creds(credentials.NewTLS(c))
	f.server = grpc.NewServer(opt)
	loggregator_v2.RegisterEgressServer(f.server, f)

	if err := f.server.Serve(lis); err != nil {
		// TODO
		// log.Fatalf("could not start gRPC server: %s", err)
	}
}

func (f *fakeEventProducer) stop() bool {
	if f.server == nil {
		return false
	}

	f.server.Stop()
	f.server = nil
	return true
}

func (f *fakeEventProducer) actualReq() *loggregator_v2.EgressBatchRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.actualReq_
}

func (f *fakeEventProducer) connectionAttempts() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.connectionAttempts_
}

func newServerMutualTLSConfig() (*tls.Config, error) {
	certFile := "fixtures/testserver.crt"
	keyFile := "fixtures/testserver.key"
	caCertFile := "fixtures/ca.crt"

	tlsCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load keypair: %s", err)
	}

	certBytes, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read ca cert file: %s", err)
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(certBytes); !ok {
		return nil, errors.New("unable to load ca cert file")
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS12,
		Certificates:       []tls.Certificate{tlsCert},
		ClientCAs:          caCertPool,
		ClientAuth:         tls.RequireAndVerifyClientCert,
	}

	return tlsConfig, nil
}
