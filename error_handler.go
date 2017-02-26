package fluent

import (
	"encoding/json"
	"io"
)

type ErrorHandler interface {
	HandleError(error, []*Message) error
}

type ErrorHandlerFunc func(error, []*Message) error

func (f ErrorHandlerFunc) HandleError(err error, messages []*Message) error {
	return f(err, messages)
}

type FallbackHandler struct {
	logger *Logger
}

func NewFallbackHandler(logger *Logger) *FallbackHandler {
	return &FallbackHandler{
		logger: logger,
	}
}

func (h *FallbackHandler) HandleError(err error, messages []*Message) error {
	return h.logger.writeWithBreaker(messages)
}

type FallbackIOHandler struct {
	io io.Writer
}

func NewFallbackIOHandler(io io.Writer) *FallbackIOHandler {
	return &FallbackIOHandler{io: io}
}

func (h *FallbackIOHandler) HandleError(err error, messages []*Message) error {
	for _, m := range messages {
		data, e := json.Marshal(m.Slice())
		if e != nil {
			return err
		}
		_, e = h.io.Write(append(data, "\n"...))
		if e != nil {
			return err
		}
	}
	return nil
}
