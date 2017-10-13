package refnozzle_test

import (
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"code.cloudfoundry.org/refnozzle"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Writer", func() {
	It("reports events", func() {
		consumer := newFakeEventConsumer()
		consumer.start()
		defer consumer.stop()

		buf := newSpyEnvelopeBuffer()
		buf.readEnvelope = &loggregator_v2.Envelope{
			SourceId: "a",
			Message: &loggregator_v2.Envelope_Event{
				Event: &loggregator_v2.Event{
					Title: "some-title",
					Body:  "some-body",
				},
			},
		}
		w := refnozzle.NewWriter(buf, consumer.addr())
		go w.Start()

		expected := refnozzle.Event{
			Title: "some-title",
			Body:  "some-body",
		}
		Eventually(consumer.actualEvent).Should(Equal(expected))
	})
})
