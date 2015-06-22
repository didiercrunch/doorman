package doorman

import (
	"github.com/didiercrunch/doorman/nanomsgsubscriber"
	"github.com/didiercrunch/doorman/nsqsubscriber"
	"github.com/didiercrunch/doorman/shared"
	"github.com/didiercrunch/doorman/subscriber"
	"gopkg.in/mgo.v2/bson"
)

type Subscriber interface {
	Subscribe(doormanId bson.ObjectId, update shared.UpdateHandlerFunc) error
}

func (w *Doorman) subscribe(sub Subscriber) error {
	return sub.Subscribe(w.Id, w.Update)
}

func (w *Doorman) NSQSubscriber(NSQLookupdURl string) error {
	sub := &nsqsubscriber.NSQSubscriber{NSQLookupdURl}
	return w.subscribe(sub)
}

func (w *Doorman) NanoMsgSubscriber(NanoMsgUrlLookupdURl string) error {
	sub := &nanomsgsubscriber.NanoMsgSubscriber{NanoMsgUrlLookupdURl}
	return w.subscribe(sub)
}

func (w *Doorman) Subscriber(serverUrl string) error {
	sub := &subscriber.Subscriber{serverUrl}
	return w.subscribe(sub)
}
