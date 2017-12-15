package fluent

import (
	"sync"
)

type buffer struct {
	buf   []*Message
	mu    sync.Mutex
	Dirty chan struct{}
}

func newBuffer() buffer {
	return buffer{
		Dirty: make(chan struct{}),
	}
}

func (buffer *buffer) Add(message *Message) {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()

	buffer.buf = append(buffer.buf, message)
	go func() {
		buffer.Dirty <- struct{}{}
	}()
}

func (buffer *buffer) Remove() []*Message {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()

	if len(buffer.buf) == 0 {
		return nil
	}

	m := buffer.buf
	buffer.buf = buffer.buf[:0]
	return m
}

func (buffer *buffer) Back(messages []*Message) {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()

	buffer.buf = append(buffer.buf, messages...)
}
