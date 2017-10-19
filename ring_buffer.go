package refnozzle

import (
	"code.cloudfoundry.org/go-diodes"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
)

type RingBuffer struct {
	d *diodes.Poller
}

func NewRingBuffer(size int, alerter diodes.Alerter) *RingBuffer {
	return &RingBuffer{
		d: diodes.NewPoller(diodes.NewOneToOne(size, alerter)),
	}
}

func (d *RingBuffer) Write(data *loggregator_v2.Envelope) {
	d.d.Set(diodes.GenericDataType(data))
}

func (d *RingBuffer) Read() *loggregator_v2.Envelope {
	data := d.d.Next()
	return (*loggregator_v2.Envelope)(data)
}
