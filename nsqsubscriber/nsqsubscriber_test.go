package nsqsubscriber

import (
	"testing"

	"github.com/bitly/go-nsq"
)

func TestNSQMessage(t *testing.T) {
	m := &nsq.Message{Body: []byte("hello world"), Timestamp: 10}
	nsqM := &NSQMessage{m}
	if string(nsqM.Body()) != "hello world" {
		t.Error()
	}

	if nsqM.Timestamp() != 10 {
		t.Error()
	}
}
