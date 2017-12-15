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
	defaultAddress       = "127.0.0.1:24224"
	defaultFlushInterval = 5 * time.Second
)

type Config struct {
	Address           string
	ConnectionTimeout time.Duration
	FlushInterval     time.Duration
}

func withDefaultConfig(c Config) Config {
	if c.Address == "" {
		c.Address = defaultAddress
	}
	if c.FlushInterval == 0 {
		c.FlushInterval = defaultFlushInterval
	}
	return c
}

type Logger struct {
	conf Config
	conn io.WriteCloser
	mu   sync.Mutex
	buf  buffer
	wg   sync.WaitGroup
	done chan struct{}
}

func NewLogger(c Config) (*Logger, error) {
	logger := &Logger{
		conf: withDefaultConfig(c),
		buf:  newBuffer(),
		done: make(chan struct{}),
	}
	if err := logger.connect(); err != nil {
		return nil, err
	}
	logger.start()
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
	logger.buf.Add(buf.Bytes())
	return nil
}

func (logger *Logger) Close() error {
	logger.stop()
	return logger.disconnect()
}

func (logger *Logger) connect() error {
	logger.mu.Lock()
	defer logger.mu.Unlock()

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
	logger.mu.Lock()
	defer logger.mu.Unlock()

	if logger.conn == nil {
		return nil
	}
	err := logger.conn.Close()
	logger.conn = nil
	return err
}

const maxWriteAttempts = 3

func (logger *Logger) send() error {
	data := logger.buf.Remove()
	if len(data) == 0 {
		return nil
	}

	var err error
	for i := 0; i < maxWriteAttempts; i++ {
		err = logger.connect()
		if err == nil {
			_, err := logger.conn.Write(data)
			if err == nil {
				break
			}
		}
		logger.disconnect()
	}
	if err != nil {
		logger.buf.Back(data)
	}
	return err
}

func (logger *Logger) start() {
	ticker := time.NewTicker(logger.conf.FlushInterval)
	logger.wg.Add(1)
	go func() {
		defer logger.wg.Done()
		for {
			select {
			case <-logger.done:
				logger.send()
				return
			case <-logger.buf.Dirty:
			case <-ticker.C:
			}
			logger.send()
		}
	}()
}

func (logger *Logger) stop() {
	close(logger.done)
	logger.wg.Wait()
}
