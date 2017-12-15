package fluent

import (
	"sync"
)

type buffer struct {
	new     []*Message
	pending []*Message
	mu      sync.Mutex
	Dirty   chan struct{}
}

func newBuffer() buffer {
	return buffer{
		Dirty: make(chan struct{}),
	}
}

func (buffer *buffer) Add(message *Message) {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()

	buffer.new = append(buffer.new, message)
	if len(buffer.new) == 1 {
		go func() {
			buffer.Dirty <- struct{}{}
		}()
	}
}

func (buffer *buffer) Remove() []*Message {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()

	if len(buffer.new) == 0 && len(buffer.pending) == 0 {
		return nil
	}

	var messages []*Message
	messages = append(messages, buffer.new...)
	messages = append(messages, buffer.pending...)

	buffer.new = buffer.new[:0]
	buffer.pending = buffer.pending[:0]
	return messages
}

func (buffer *buffer) Back(messages []*Message) {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()

	buffer.pending = append(buffer.pending, messages...)
}
