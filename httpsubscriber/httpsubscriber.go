package httpsubscriber

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/didiercrunch/doorman/shared"
	"gopkg.in/mgo.v2/bson"
)

type HttpSubscriber struct {
	Url      string
	HartBeat time.Duration
}

func (s *HttpSubscriber) GetDoormanUpdater() (*shared.DoormanUpdater, error) {
	resp, err := http.Get(s.Url)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, errors.New("bad http status when GETting update, " + resp.Status)
	}
	defer resp.Body.Close()
	de := json.NewDecoder(resp.Body)
	ret := new(shared.DoormanUpdater)
	err = de.Decode(ret)
	return ret, err
}

func (s *HttpSubscriber) callUpdateHandlerFunction(f shared.UpdateHandlerFunc, data []byte) error {
	wu := &shared.DoormanUpdater{}
	if err := json.Unmarshal(data, wu); err != nil {
		return err
	} else {
		return f(wu)
	}
}

func (s *HttpSubscriber) Subscribe(abtestId bson.ObjectId, update shared.UpdateHandlerFunc) error {
	go func() {
		c := time.Tick(s.HartBeat)
		for _ = range c {
			if du, err := s.GetDoormanUpdater(); err != nil {
				log.Printf("error retrieving doorman %v \n%v\n", abtestId, err)
			} else {
				update(du)
			}
		}
	}()
	return nil
}
