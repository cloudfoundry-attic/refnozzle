package refnozzle_test

import (
	"context"

	"code.cloudfoundry.org/diodes"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"code.cloudfoundry.org/refnozzle"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RefNozzle", func() {
	It("reports received events", func() {
		producer, err := newFakeEventProducer()
		Expect(err).NotTo(HaveOccurred())
		go producer.start()
		defer producer.stop()

		consumer := newFakeEventConsumer()
		consumer.start()
		defer consumer.stop()

		buf := refnozzle.NewRingBuffer(5, diodes.AlertFunc(func(int) {}))
		tlsConf, err := refnozzle.NewClientMutualTLSConfig(
			"fixtures/refnozzle.crt",
			"fixtures/refnozzle.key",
			"fixtures/ca.crt",
			"testserver",
		)
		Expect(err).NotTo(HaveOccurred())
		w := refnozzle.NewWriter(
			buf,
			consumer.addr(),
		)
		go w.Start()

		eventsOnly := &loggregator_v2.EgressBatchRequest{
			ShardId: "some-id",
			Selectors: []*loggregator_v2.Selector{
				{
					Message: &loggregator_v2.Selector_Event{
						Event: &loggregator_v2.EventSelector{},
					},
				},
			},
		}
		c := refnozzle.NewEnvelopeStreamConnector(
			producer.addr,
			tlsConf,
		)

		rx := c.Stream(context.Background(), eventsOnly)

		r := refnozzle.NewRepeater(
			buf,
			rx,
		)
		go r.Start()

		Eventually(producer.actualReq).Should(Equal(eventsOnly))
		expectedEvent := refnozzle.Event{
			Title: "event-name",
			Body:  "event-body",
		}
		Eventually(consumer.actualEvent).Should(Equal(expectedEvent))
	})
})
