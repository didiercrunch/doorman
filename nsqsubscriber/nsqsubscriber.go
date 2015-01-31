package nsqsubscriber

import (
	"encoding/json"

	"code.google.com/p/go-uuid/uuid"
	"github.com/bitly/go-nsq"
	"github.com/didiercrunch/doorman/shared"
	"gopkg.in/mgo.v2/bson"
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

func (sub *NSQSubscriber) Subscribe(doormanId bson.ObjectId, update shared.UpdateHandlerFunc) error {
	config := nsq.NewConfig()
	q, err := nsq.NewConsumer(doormanId.Hex(), UUID, config)
	if err != nil {
		return err
	}
	q.AddHandler(nsq.HandlerFunc(toNSQHandlerFunc(update)))
	if err := q.ConnectToNSQLookupd(sub.NSQLookupURL); err != nil {
		return err
	}
	return nil
}
