package fluent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/vmihailenco/msgpack.v2"
)

type ErrorHandler interface {
	HandleError(error, []byte) error
}

type ErrorHandlerFunc func(error, []byte) error

func (f ErrorHandlerFunc) HandleError(err error, data []byte) error {
	return f(err, data)
}

type FallbackHandler struct {
	logger *Logger
}

func NewFallbackHandler(logger *Logger) *FallbackHandler {
	return &FallbackHandler{
		logger: logger,
	}
}

func (h *FallbackHandler) HandleError(_ error, data []byte) error {
	return h.logger.writeWithBreaker(data)
}

type FallbackJSONHandler struct {
	io io.Writer
}

func NewFallbackJSONHandler(io io.Writer) *FallbackJSONHandler {
	return &FallbackJSONHandler{io: io}
}

func (h *FallbackJSONHandler) HandleError(_ error, data []byte) error {
	r := bytes.NewReader(data)
	d := msgpack.NewDecoder(r)

	for r.Len() > 0 {
		var tag string
		var timestamp uint64
		iRecord := map[interface{}]interface{}{}

		message := []interface{}{&tag, &timestamp, &iRecord}
		if err := d.Decode(&message); err != nil {
			return err
		}
		record := h.castMap(iRecord)

		str, err := json.Marshal([]interface{}{tag, timestamp, record})
		if err != nil {
			return err
		}
		_, err = h.io.Write(append(str, "\n"...))
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *FallbackJSONHandler) castMap(in map[interface{}]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range in {
		res[fmt.Sprintf("%v", k)] = h.castValue(v)
	}
	return res
}

func (h *FallbackJSONHandler) castValue(iv interface{}) interface{} {
	switch val := iv.(type) {
	case map[interface{}]interface{}:
		return h.castMap(val)
	case []interface{}:
		ia := make([]interface{}, len(val))
		for i, v := range val {
			ia[i] = h.castValue(v)
		}
		return ia
	default:
		return iv
	}
}
