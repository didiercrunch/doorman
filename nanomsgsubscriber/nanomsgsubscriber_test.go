package nanomsgsubscriber

import (
	"encoding/json"
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

	if new(NanoMsgSubscriber).callUpdateHandlerFunction(f, data); err != nil {
		t.Error(err)
	}
}
