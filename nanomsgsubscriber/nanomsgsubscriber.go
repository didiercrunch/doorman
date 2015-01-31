package nanomsgsubscriber

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/didiercrunch/doorman/shared"
	"github.com/gdamore/mangos"
	"github.com/gdamore/mangos/protocol/sub"
	"github.com/gdamore/mangos/transport/ipc"
	"github.com/gdamore/mangos/transport/tcp"
	"gopkg.in/mgo.v2/bson"
)

type NanoMsgSubscriber struct {
	Url string
}

func (s *NanoMsgSubscriber) callUpdateHandlerFunction(f shared.UpdateHandlerFunc, data []byte) error {
	wu := &shared.DoormanUpdater{}
	if err := json.Unmarshal(data, wu); err != nil {
		return err
	} else {
		return f(wu)
	}
}

func (s *NanoMsgSubscriber) Subscribe(abtestId bson.ObjectId, update shared.UpdateHandlerFunc) error {
	var sock mangos.Socket
	var err error
	var msg []byte

	if sock, err = sub.NewSocket(); err != nil {
		return errors.New("can't get new sub socket: " + err.Error())
	}
	sock.AddTransport(ipc.NewTransport())
	sock.AddTransport(tcp.NewTransport())
	if err = sock.Dial(s.Url); err != nil {
		return errors.New("can't dial on sub socket: " + err.Error())
	}
	// Empty byte array effectively subscribes to everything
	if err = sock.SetOption(mangos.OptionSubscribe, []byte("")); err != nil {
		return errors.New("cannot subscribe: " + err.Error())
	}
	go func(s *NanoMsgSubscriber, sock mangos.Socket) {
		for {
			if msg, err = sock.Recv(); err != nil {
				err := errors.New("Cannot recv: " + err.Error())
				log.Println(err)
			} else if err := s.callUpdateHandlerFunction(update, msg); err != nil {
				log.Println("cannot update abtest with received data: ", err)
			}
		}
	}(s, sock)

	return nil

}
