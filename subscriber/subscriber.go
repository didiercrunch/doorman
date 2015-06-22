package subscriber

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/didiercrunch/doorman/httpsubscriber"
	"github.com/didiercrunch/doorman/nanomsgsubscriber"
	"github.com/didiercrunch/doorman/shared"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"time"
)

type Subscriber struct {
	URL string
}

type subscriber interface {
	Subscribe(abtestId bson.ObjectId, update shared.UpdateHandlerFunc) error
}

type ServerSpecification struct {
	HostName     string            `json:"hostname"`
	Port         int               `json:"port"`
	MessageQueue string            `json:"message_queue"`
	NanoMsg      map[string]string `json:"nano_msg"`
}

func (s *Subscriber) getServerSpecification() (*ServerSpecification, error) {
	url := s.URL + "/api/server"
	fmt.Println(">>>", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()
	ret := new(ServerSpecification)
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(ret)
	fmt.Println("ok!")
	return ret, err
}

func (sub *Subscriber) GetSubsciber(serverSpec *ServerSpecification, doormanId bson.ObjectId) subscriber {
	switch serverSpec.MessageQueue {
	case "nanomsg":
		return &nanomsgsubscriber.NanoMsgSubscriber{serverSpec.NanoMsg["url"]}
	}
	fmt.Println("shit!", serverSpec)
	return &httpsubscriber.HttpSubscriber{sub.getDoormanStatusUrl(doormanId), time.Second * 5}
}

func (sub *Subscriber) getDoormanStatusUrl(doormanId bson.ObjectId) string {
	return sub.URL + "/api/doormen/" + doormanId.Hex() + "/status"
}

func (sub *Subscriber) SetInitialState(doormanId bson.ObjectId, update shared.UpdateHandlerFunc) error {
	httpSUbscriber := &httpsubscriber.HttpSubscriber{Url: sub.getDoormanStatusUrl(doormanId)}
	if du, err := httpSUbscriber.GetDoormanUpdater(); err != nil {
		return err
	} else {
		return update(du)
	}

}

func startHartBeat() error {
	return nil
}

func (sub *Subscriber) Subscribe(doormanId bson.ObjectId, update shared.UpdateHandlerFunc) error {
	spec, err := sub.getServerSpecification()
	if err != nil {
		return err
	}
	if err = sub.SetInitialState(doormanId, update); err != nil {
		return err
	}
	subscriber := sub.GetSubsciber(spec, doormanId)
	return subscriber.Subscribe(doormanId, update)
}
