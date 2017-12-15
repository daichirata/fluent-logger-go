package fluent

import (
	"bytes"
	"sync"
)

type Message struct {
	buf *bytes.Buffer
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
	return &Message{
		buf: bytes.NewBuffer([]byte{}),
	}
}
