package nsqsubscriber

import "github.com/bitly/go-nsq"

type NSQMessage struct {
	*nsq.Message
}

func (m *NSQMessage) Body() []byte {
	return m.Message.Body
}

func (m *NSQMessage) Timestamp() int64 {
	return m.Message.Timestamp
}
