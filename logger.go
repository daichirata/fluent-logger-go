package fluent

import (
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"bytes"

	"gopkg.in/vmihailenco/msgpack.v2"
)

var (
	defaultAddress = "127.0.0.1:24224"
)

type Config struct {
	Address           string
	ConnectionTimeout time.Duration
}

func withDefaultConfig(c Config) Config {
	if c.Address == "" {
		c.Address = defaultAddress
	}
	return c
}

type Logger struct {
	conf Config
	conn io.WriteCloser
	bmu  sync.Mutex
	cmu  sync.Mutex
	buf  []byte
}

func NewLogger(c Config) (*Logger, error) {
	logger := &Logger{
		conf: withDefaultConfig(c),
		buf:  []byte{},
	}
	if err := logger.connect(); err != nil {
		return nil, err
	}
	return logger, nil
}

func (logger *Logger) Post(tag string, obj interface{}) error {
	return logger.PostWithTime(tag, time.Now(), obj)
}

func (logger *Logger) PostWithTime(tag string, t time.Time, obj interface{}) error {
	record := []interface{}{
		tag,
		t.Unix(),
		obj,
	}

	buf := bytes.NewBuffer([]byte{})
	enc := msgpack.NewEncoder(buf)
	if err := enc.Encode(record); err != nil {
		return err
	}
	raw := buf.Bytes()

	logger.bmu.Lock()
	logger.buf = append(logger.buf, raw...)
	logger.bmu.Unlock()

	return logger.send()
}

func (logger *Logger) Close() error {
	return logger.disconnect()
}

func (logger *Logger) connect() error {
	logger.cmu.Lock()
	defer logger.cmu.Unlock()

	if logger.conn != nil {
		return nil
	}
	var err error
	if strings.HasPrefix(logger.conf.Address, "unix:/") {
		logger.conn, err = net.DialTimeout(
			"unix",
			logger.conf.Address[5:],
			logger.conf.ConnectionTimeout,
		)
	} else {
		logger.conn, err = net.DialTimeout(
			"tcp",
			logger.conf.Address,
			logger.conf.ConnectionTimeout,
		)
	}
	return err
}

func (logger *Logger) disconnect() error {
	logger.cmu.Lock()
	defer logger.cmu.Unlock()

	if logger.conn == nil {
		return nil
	}
	err := logger.conn.Close()
	logger.conn = nil
	return err
}

const maxWriteAttempts = 3

func (logger *Logger) send() error {
	logger.bmu.Lock()
	defer logger.bmu.Unlock()

	data := logger.buf
	if len(data) == 0 {
		return nil
	}
	var err error
	for i := 0; i < maxWriteAttempts; i++ {
		err = logger.connect()
		if err == nil {
			_, err := logger.conn.Write(data)
			if err == nil {
				logger.buf = logger.buf[:0]
				break
			}
		}
		logger.disconnect()
	}

	return err
}
