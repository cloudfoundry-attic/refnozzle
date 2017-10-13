package refnozzle

import (
	"log"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
)

type Repeater struct {
	buf    EnvelopeBufferWriter
	stream Stream
}

type EnvelopeBufferWriter interface {
	Write(*loggregator_v2.Envelope)
}

type Stream interface {
	Receive() []*loggregator_v2.Envelope
}

func NewRepeater(
	buf EnvelopeBufferWriter,
	s Stream,
) *Repeater {
	return &Repeater{
		buf:    buf,
		stream: s,
	}
}

func (r *Repeater) Start() {
	for {
		envs := r.stream.Receive()
		for _, e := range envs {
			log.Printf("Received event envelope: %+v", e)
			r.buf.Write(e)
		}
	}
}
