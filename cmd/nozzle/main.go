package main

import (
	"flag"
	"log"

	"code.cloudfoundry.org/diodes"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"code.cloudfoundry.org/refnozzle"
)

func main() {
	caPath := flag.String(
		"ca",
		"certs/ca.crt",
		"Path to the CA cert used for mutual TLS",
	)
	certPath := flag.String(
		"cert",
		"certs/refnozzle.crt",
		"Path to the client cert",
	)
	keyPath := flag.String(
		"key",
		"certs/refnozzle.key",
		"Path to the client key",
	)
	src := flag.String(
		"events-src",
		"",
		"URL from which to retrieve events",
	)
	dest := flag.String(
		"report-url",
		"",
		"URL to which to report events",
	)
	shardID := flag.String(
		"shard-id",
		"refnozzle",
		"Unique ID that identifies this nozzle",
	)
	flag.Parse()

	tlsConfig, err := refnozzle.NewClientMutualTLSConfig(
		*certPath,
		*keyPath,
		*caPath,
		"rlp",
	)
	if err != nil {
		log.Fatalf("failed to create mutual TLS config: %v", err)
	}

	buf := refnozzle.NewRingBuffer(10000, diodes.AlertFunc(func(n int) {
		log.Println("dropped %d envelopes. Consider scaling the number of"+
			"nozzle instances up, or the downstream consumer needs to be faster...", n)
	}))

	w := refnozzle.NewWriter(
		buf,
		*dest,
	)
	go w.Start()

	eventsOnly := &loggregator_v2.EgressBatchRequest{
		ShardId: *shardID,
		Selectors: []*loggregator_v2.Selector{
			{
				Message: &loggregator_v2.Selector_Event{
					Event: &loggregator_v2.EventSelector{},
				},
			},
		},
	}
	c := refnozzle.NewConnector(
		eventsOnly,
		*src,
		tlsConfig,
	)
	r := refnozzle.NewRepeater(
		buf,
		c,
	)
	r.Start()
}
