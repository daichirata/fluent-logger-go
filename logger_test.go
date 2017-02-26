package fluent_test

import (
	"io"
	"net"
	"testing"
	"time"

	fluent "github.com/daichirata/fluent-logger-go"
	"github.com/stretchr/testify/assert"
	"gopkg.in/vmihailenco/msgpack.v2"
)

func newTCPListener() net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	return l
}

func newUnixListener() net.Listener {
	l, err := net.Listen("unix", "/tmp/fluent-logger-go.sock")
	if err != nil {
		panic(err)
	}
	return l
}

func receive(l net.Listener, ch chan []byte) {
	conn, err := l.Accept()
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		ch <- buf[:n]
	}
}

func TestNewLoggerConnectWithTCP(t *testing.T) {
	listener := newTCPListener()
	addr := listener.Addr().String()

	_, err := fluent.NewLogger(fluent.Config{
		Address: addr,
	})
	if err != nil {
		t.Fatal(err)
	}

	listener.Close()
	_, err = fluent.NewLogger(fluent.Config{
		Address: addr,
	})
	if err == nil {
		t.Fatal(err)
	}
}

func TestNewLoggerConnectWithUnixDomainSocket(t *testing.T) {
	listener := newUnixListener()
	addr := "unix:/" + listener.Addr().String()

	_, err := fluent.NewLogger(fluent.Config{
		Address: addr,
	})
	if err != nil {
		t.Fatal(err)
	}

	listener.Close()
	_, err = fluent.NewLogger(fluent.Config{
		Address: addr,
	})
	if err == nil {
		t.Fatal(err)
	}
}

func TestPost(t *testing.T) {
	listener := newTCPListener()
	defer listener.Close()
	ch := make(chan []byte)
	go receive(listener, ch)

	logger, err := fluent.NewLogger(fluent.Config{
		Address: listener.Addr().String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	if err := logger.Post("test.tag", map[string]string{"key": "value"}); err != nil {
		t.Fatal(err)
	}
	buf := <-ch
	row := []interface{}{}
	if err := msgpack.Unmarshal(buf, &row); err != nil {
		t.Fatal(err)
	}
	assert.Len(t, row, 3)
	assert.Equal(t, "test.tag", row[0])
	assert.IsType(t, uint64(0), row[1])
	assert.Equal(t, map[interface{}]interface{}{"key": "value"}, row[2])
}

func TestPostWithTime(t *testing.T) {
	listener := newTCPListener()
	defer listener.Close()
	ch := make(chan []byte)
	go receive(listener, ch)

	logger, err := fluent.NewLogger(fluent.Config{
		Address: listener.Addr().String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	now := time.Unix(time.Now().Unix(), 0)
	if err := logger.PostWithTime("test.tag", now, map[string]string{"key": "value"}); err != nil {
		t.Fatal(err)
	}
	buf := <-ch
	row := []interface{}{}
	if err := msgpack.Unmarshal(buf, &row); err != nil {
		t.Fatal(err)
	}
	assert.Len(t, row, 3)
	assert.Equal(t, "test.tag", row[0])
	assert.Equal(t, uint64(now.Unix()), row[1])
	assert.Equal(t, map[interface{}]interface{}{"key": "value"}, row[2])
}
