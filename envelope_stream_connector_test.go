package refnozzle_test

import (
	"context"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"code.cloudfoundry.org/refnozzle"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Connector", func() {
	It("initiates a connection to receive envelopes", func() {
		producer, err := newFakeEventProducer()
		Expect(err).NotTo(HaveOccurred())
		go producer.start()
		defer producer.stop()
		tlsConf, err := refnozzle.NewClientMutualTLSConfig(
			"fixtures/refnozzle.crt",
			"fixtures/refnozzle.key",
			"fixtures/ca.crt",
			"testserver",
		)
		Expect(err).NotTo(HaveOccurred())

		req := &loggregator_v2.EgressBatchRequest{ShardId: "some-id"}
		c := refnozzle.NewEnvelopeStreamConnector(
			producer.addr,
			tlsConf,
		)

		rx := c.Stream(context.Background(), req)

		Expect(len(rx())).NotTo(BeZero())
		Expect(producer.actualReq()).To(Equal(req))
	})

	It("reconnects if the stream fails", func() {
		producer, err := newFakeEventProducer()
		Expect(err).NotTo(HaveOccurred())

		// Producer will grab a port on start. When the producer is restarted,
		// it will grab the same port.
		go producer.start()

		tlsConf, err := refnozzle.NewClientMutualTLSConfig(
			"fixtures/refnozzle.crt",
			"fixtures/refnozzle.key",
			"fixtures/ca.crt",
			"testserver",
		)
		Expect(err).NotTo(HaveOccurred())

		c := refnozzle.NewEnvelopeStreamConnector(
			producer.addr,
			tlsConf,
		)

		go func() {
			rx := c.Stream(context.Background(), &loggregator_v2.EgressBatchRequest{})
			for {
				rx()
			}
		}()

		Eventually(producer.connectionAttempts).Should(Equal(1))
		producer.stop()
		go producer.start()

		// Reconnect after killing the server.
		Eventually(producer.connectionAttempts, 5).Should(Equal(2))

		// Ensure we don't create new connections when we don't need to.
		Consistently(producer.connectionAttempts).Should(Equal(2))
	})
})
