package doorman

import (
	"github.com/didiercrunch/doorman/nanomsgsubscriber"
	"github.com/didiercrunch/doorman/nsqsubscriber"
	"github.com/didiercrunch/doorman/shared"
	"gopkg.in/mgo.v2/bson"
)

type Subscriber interface {
	Subscribe(doormanId bson.ObjectId, update shared.UpdateHandlerFunc) error
}

func (w *Doorman) Subscribe(sub Subscriber) error {
	return sub.Subscribe(w.Id, w.Update)
}

func (w *Doorman) NSQSubscriber(NSQLookupdURl string) error {
	sub := &nsqsubscriber.NSQSubscriber{NSQLookupdURl}
	return w.Subscribe(sub)
}

func (w *Doorman) NanoMsgSubscriber(NanoMsgUrlLookupdURl string) error {
	sub := &nanomsgsubscriber.NanoMsgSubscriber{NanoMsgUrlLookupdURl}
	return w.Subscribe(sub)
}
