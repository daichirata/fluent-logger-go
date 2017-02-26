package fluent

import (
	"bytes"
	"sync"
	"time"

	"gopkg.in/vmihailenco/msgpack.v2"
)

var messagePool = sync.Pool{
	New: func() interface{} {
		buf := bytes.NewBuffer([]byte{})
		return &Message{
			buf: buf,
			enc: msgpack.NewEncoder(buf),
		}
	},
}

func getMessage() *Message {
	message := messagePool.Get().(*Message)
	message.buf.Reset()
	return message
}

func putMessage(message *Message) {
	messagePool.Put(message)
}

type Message struct {
	Tag    string
	Time   time.Time
	Record interface{}

	buf *bytes.Buffer
	enc *msgpack.Encoder
}

func newMessage(tag string, t time.Time, obj interface{}) (*Message, error) {
	message := getMessage()
	message.Tag = tag
	message.Time = t
	message.Record = obj
	if err := message.enc.Encode(message.Slice()); err != nil {
		return nil, err
	}
	return message, nil
}

func (m *Message) Bytes() []byte {
	return m.buf.Bytes()
}

func (m *Message) Slice() []interface{} {
	return []interface{}{m.Tag, m.Time.Unix(), m.Record}
}
