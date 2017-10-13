package refnozzle_test

import (
	"sync"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"code.cloudfoundry.org/refnozzle"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Repeater", func() {
	It("writes envelopes from the stream to the buffer", func() {
		buf := newSpyEnvelopeBuffer()
		stream := func() []*loggregator_v2.Envelope {
			return []*loggregator_v2.Envelope{
				{SourceId: "a"},
				{SourceId: "b"},
			}
		}
		r := refnozzle.NewRepeater(buf, stream)
		go r.Start()

		Eventually(func() int {
			return len(buf.writeEnvelopes())
		}).Should(BeNumerically(">=", 2))
	})
})

type spyEnvelopeBuffer struct {
	readEnvelope *loggregator_v2.Envelope

	mu              sync.Mutex
	writeEnvelopes_ []*loggregator_v2.Envelope
}

func newSpyEnvelopeBuffer() *spyEnvelopeBuffer {
	return &spyEnvelopeBuffer{}
}

func (s *spyEnvelopeBuffer) Read() *loggregator_v2.Envelope {
	return s.readEnvelope
}

func (s *spyEnvelopeBuffer) Write(e *loggregator_v2.Envelope) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.writeEnvelopes_ = append(s.writeEnvelopes_, e)
}

func (s *spyEnvelopeBuffer) writeEnvelopes() []*loggregator_v2.Envelope {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.writeEnvelopes_
}
