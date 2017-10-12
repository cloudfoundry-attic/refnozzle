package refnozzle

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"code.cloudfoundry.org/refnozzle/rpc/loggregator_v2"
)

type Writer struct {
	addr string
	buf  EnvelopeBufferReader
}

type EnvelopeBufferReader interface {
	Read() *loggregator_v2.Envelope
}

func NewWriter(buf EnvelopeBufferReader, addr string) *Writer {
	return &Writer{
		addr: addr,
		buf:  buf,
	}
}

func (w *Writer) Start() {
	for {
		e := w.buf.Read()
		data := Event{
			Title: e.GetEvent().GetTitle(),
			Body:  e.GetEvent().GetBody(),
		}
		jd, err := json.Marshal(data)
		if err != nil {
			// We are marshalling a normal struct and this error should
			// never occur.
			log.Panic(err)
		}

		resp, err := http.Post(w.addr, "application/json", bytes.NewReader(jd))
		if err != nil {
			log.Println("failed to POST")
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("failed to read response body: %s", err)
				continue
			}
			log.Printf("failed to POST: %s", data)
		}
	}
}

type Event struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}
