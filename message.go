package fluent

import (
	"bytes"
	"sync"

	"gopkg.in/vmihailenco/msgpack.v2"
)

type Message struct {
	buf *bytes.Buffer
	enc *msgpack.Encoder
}

var messagePool = sync.Pool{
	New: func() interface{} {
		return newMessage()
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

func newMessage() *Message {
	buf := bytes.NewBuffer([]byte{})
	return &Message{
		buf: buf,
		enc: msgpack.NewEncoder(buf),
	}
}
