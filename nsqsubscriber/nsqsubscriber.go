package nsqsubscriber

import (
	"encoding/json"

	"github.com/pborman/uuid"
	"github.com/bitly/go-nsq"
	"github.com/didiercrunch/doorman/shared"
)

var UUID = uuid.New()

type NSQSubscriber struct {
	NSQLookupURL string
}

func toNSQHandlerFunc(update shared.UpdateHandlerFunc) nsq.HandlerFunc {
	return func(message *nsq.Message) error {
		wu := &shared.DoormanUpdater{}
		if err := json.Unmarshal(message.Body, wu); err != nil {
			return err
		}
		wu.Timestamp = message.Timestamp
		return update(wu)
	}
}

func (sub *NSQSubscriber) Subscribe(doormanId string, update shared.UpdateHandlerFunc) error {
	config := nsq.NewConfig()
	q, err := nsq.NewConsumer(doormanId, UUID, config)
	if err != nil {
		return err
	}
	q.AddHandler(nsq.HandlerFunc(toNSQHandlerFunc(update)))
	if err := q.ConnectToNSQLookupd(sub.NSQLookupURL); err != nil {
		return err
	}
	return nil
}
