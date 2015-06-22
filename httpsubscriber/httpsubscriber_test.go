package httpsubscriber

import (
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/didiercrunch/doorman/shared"
)

func TestCallUpdateHandlerFunction(t *testing.T) {
	f := func(m *shared.DoormanUpdater) error {
		if m.Id != "foo" {
			t.Error("bad doorman id", m.Id)
		}
		return nil
	}

	data, err := json.Marshal(&shared.DoormanUpdater{Id: "foo"})
	if err != nil {
		t.Error(err)
		return
	}

	if new(HttpSubscriber).callUpdateHandlerFunction(f, data); err != nil {
		t.Error(err)
	}
}

func MockEndpoint(w http.ResponseWriter, r *http.Request) {
	du := &shared.DoormanUpdater{
		"b64",
		897987,
		[]*big.Rat{big.NewRat(1, 4), big.NewRat(3, 4)},
	}
	enc := json.NewEncoder(w)
	if err := enc.Encode(du); err != nil {
		panic(err)
	}

}

func TestGetDoormanUpdater(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(MockEndpoint))
	defer ts.Close()

	s := HttpSubscriber{Url: ts.URL}
	if d, err := s.GetDoormanUpdater(); err != nil {
		t.Error(err)
	} else if d.Id != "b64" {
		t.Error()
	}

}
