package refnozzle

import (
	"log"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
)

type Repeater struct {
	buf    EnvelopeBufferWriter
	stream EnvelopeStream
}

type EnvelopeBufferWriter interface {
	Write(*loggregator_v2.Envelope)
}

func NewRepeater(
	buf EnvelopeBufferWriter,
	s EnvelopeStream,
) *Repeater {
	return &Repeater{
		buf:    buf,
		stream: s,
	}
}

func (r *Repeater) Start() {
	for {
		envs := r.stream()
		for _, e := range envs {
			log.Printf("Received event envelope: %+v", e)
			r.buf.Write(e)
		}
	}
}
