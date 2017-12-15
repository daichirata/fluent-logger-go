package fluent

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/rubyist/circuitbreaker"
)

var (
	defaultAddress          = "127.0.0.1:24224"
	defaultFlushInterval    = 5 * time.Second
	defaultFailureThreshold = int64(1)
)

type Config struct {
	Address           string
	ConnectionTimeout time.Duration
	FailureThreshold  int64
	FlushInterval     time.Duration
	PendingLimit      int
}

func withDefaultConfig(c Config) Config {
	if c.Address == "" {
		c.Address = defaultAddress
	}
	if c.FlushInterval == 0 {
		c.FlushInterval = defaultFlushInterval
	}
	if c.FailureThreshold == 0 {
		c.FailureThreshold = defaultFailureThreshold
	}
	return c
}

type Logger struct {
	ErrorHandler ErrorHandler

	conf    Config
	conn    io.WriteCloser
	buf     buffer
	breaker *circuit.Breaker
	mu      sync.Mutex
	wg      sync.WaitGroup
	done    chan struct{}
}

func NewLogger(c Config) (*Logger, error) {
	conf := withDefaultConfig(c)
	logger := &Logger{
		conf:    conf,
		buf:     newBuffer(),
		breaker: circuit.NewConsecutiveBreaker(conf.FailureThreshold),
		done:    make(chan struct{}),
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

	m := getMessage()
	if err := m.enc.Encode(record); err != nil {
		return err
	}
	logger.buf.Add(m)
	return nil
}

func (logger *Logger) Subscribe() <-chan circuit.BreakerEvent {
	return logger.breaker.Subscribe()
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

func (logger *Logger) write(data []byte) error {
	err := logger.connect()
	if err == nil {
		_, err = logger.conn.Write(data)
		if err == nil {
			return nil
		}
	}
	logger.disconnect()
	return err
}

func (logger *Logger) writeWithBreaker(data []byte) error {
	return logger.breaker.Call(func() error {
		return logger.write(data)
	}, 0)
}

func (logger *Logger) send() error {
	messages := logger.buf.Remove()
	if len(messages) == 0 {
		return nil
	}

	var data []byte
	for _, m := range messages {
		data = append(data, m.buf.Bytes()...)
	}

	err := logger.writeWithBreaker(data)
	if err != nil {
		if logger.ErrorHandler != nil && len(messages) > logger.conf.PendingLimit {
			err = logger.ErrorHandler.HandleError(err, data)
		}
		if err != nil {
			fmt.Println(err)
			logger.buf.Back(messages)
			return err
		}
	}

	for _, m := range messages {
		putMessage(m)
	}
	return nil
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
