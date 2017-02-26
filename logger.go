package fluent

import (
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
	FlushInterval     time.Duration
	PendingLimit      int
	FailureThreshold  int64
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
	// ErrorHandler handle error events. If Errorhandler is set, it is called
	// with error and message which failed to send as arguments.
	// If ErrorHandler does not return error, it means that the sending of the
	// message was successful.
	ErrorHandler ErrorHandler

	conf    Config
	conn    io.WriteCloser
	buf     buffer
	breaker *circuit.Breaker
	mu      sync.Mutex
	wg      sync.WaitGroup
	done    chan struct{}
}

// NewLogger generates a new Logger instance.
func NewLogger(c Config) (*Logger, error) {
	conf := withDefaultConfig(c)
	logger := &Logger{
		conf:    withDefaultConfig(c),
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

// Post a message to fluentd. This method returns an error if encoding to msgpack fails.
// Message posting processing is performed asynchronously by goroutine, so it will not block.
func (logger *Logger) Post(tag string, obj interface{}) error {
	return logger.PostWithTime(tag, time.Now(), obj)
}

// PostWithTime posts a message with specified time to fluentd.
func (logger *Logger) PostWithTime(tag string, t time.Time, obj interface{}) error {
	message, err := newMessage(tag, t, obj)
	if err != nil {
		return err
	}
	logger.buf.Add(message)
	return nil
}

// Subscribe returns a channel of circuit.BreakerEvents.
func (logger *Logger) Subscribe() <-chan circuit.BreakerEvent {
	return logger.breaker.Subscribe()
}

// Close will block until all messages has been sent.
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

const maxWriteAttempts = 2

func (logger *Logger) write(messages []*Message) error {
	if len(messages) == 0 {
		return nil
	}
	var data []byte
	for _, m := range messages {
		data = append(data, m.Bytes()...)
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
	return err
}

func (logger *Logger) writeWithBreaker(messages []*Message) error {
	return logger.breaker.Call(func() error {
		return logger.write(messages)
	}, 0)
}

func (logger *Logger) send() error {
	messages := logger.buf.Remove()
	err := logger.writeWithBreaker(messages)
	if err != nil {
		if logger.ErrorHandler != nil && len(messages) > logger.conf.PendingLimit {
			err = logger.ErrorHandler.HandleError(err, messages)
		}
		if err != nil {
			logger.buf.WriteBack(messages)
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
