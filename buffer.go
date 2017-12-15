package fluent

import (
	"sync"
)

type buffer struct {
	buf   []byte
	mu    sync.Mutex
	Dirty chan struct{}
}

func newBuffer() buffer {
	return buffer{
		buf:   []byte{},
		Dirty: make(chan struct{}),
	}
}

func (buffer *buffer) Add(raw []byte) {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()

	buffer.buf = append(buffer.buf, raw...)
	go func() {
		buffer.Dirty <- struct{}{}
	}()
}

func (buffer *buffer) Remove() []byte {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()

	if len(buffer.buf) == 0 {
		return nil
	}

	data := buffer.buf
	buffer.buf = buffer.buf[:0]
	return data
}

func (buffer *buffer) Back(raw []byte) {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()

	buffer.buf = append(buffer.buf, raw...)
}
