package refnozzle_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"

	"code.cloudfoundry.org/refnozzle"
)

type fakeEventConsumer struct {
	httpServer *httptest.Server

	mu           sync.Mutex
	actualEvent_ refnozzle.Event
}

func newFakeEventConsumer() *fakeEventConsumer {
	return &fakeEventConsumer{}
}

func (f *fakeEventConsumer) actualEvent() refnozzle.Event {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.actualEvent_
}

func (f *fakeEventConsumer) addr() string {
	return f.httpServer.URL
}

func (f *fakeEventConsumer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if req.Method != http.MethodPost {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	var e refnozzle.Event
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&e)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	defer req.Body.Close()
	f.actualEvent_ = e

	rw.WriteHeader(http.StatusCreated)
}

func (f *fakeEventConsumer) start() {
	f.httpServer = httptest.NewServer(f)
}

func (f *fakeEventConsumer) stop() {
	f.httpServer.Close()
}
