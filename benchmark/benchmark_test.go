package main

import (
	"io"
	"net"
	"testing"
	"time"

	f1 "github.com/daichirata/fluent-logger-go"
	f2 "github.com/fluent/fluent-logger-golang/fluent"
)

func init() {
	listener, err := net.Listen("tcp", "0.0.0.0:24224")
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				panic(err)
			}
			go func() {
				defer conn.Close()
				for {
					buf := make([]byte, 1024)
					_, err := conn.Read(buf)
					if err != nil {
						if err == io.EOF {
							break
						}
						panic(err)
					}
				}
			}()
		}
	}()
}

type Test struct {
	A string
	B string
	C int
	D int
	E string
	F string
	G []int
	H int64
	I int64
	J int64
	K int64
	L int64
	N int64
	M int64
	O int64
	P int
	Q int
	R int64
	S int
	T int
	U float64
	V string
	W int
	X int
	Y string
}

var structmsg = Test{
	A: "string",
	B: "string",
	C: 1,
	D: 2,
	E: "string",
	F: "string",
	G: []int{3, 4},
	H: time.Now().Unix(),
	I: time.Now().Unix(),
	J: time.Now().Unix(),
	K: time.Now().Unix(),
	L: time.Now().Unix(),
	N: time.Now().Unix(),
	M: time.Now().Unix(),
	O: time.Now().Unix(),
	P: 13,
	Q: 14,
	R: 15,
	S: 16,
	T: 17,
	U: 18,
	V: "string",
	W: 19,
	X: 20,
	Y: "string",
}

var mapmsg = map[string]interface{}{
	"A": "string",
	"B": "string",
	"C": 1,
	"D": 2,
	"E": "string",
	"F": "string",
	"G": []int{3, 4},
	"H": time.Now().Unix(),
	"I": time.Now().Unix(),
	"J": time.Now().Unix(),
	"K": time.Now().Unix(),
	"L": time.Now().Unix(),
	"N": time.Now().Unix(),
	"M": time.Now().Unix(),
	"O": time.Now().Unix(),
	"P": 13,
	"Q": 14,
	"R": 15,
	"S": 16,
	"T": 17,
	"U": 18,
	"V": "string",
	"W": 19,
	"X": 20,
	"Y": "string",
}

func BenchmarkStructFluentLoggerGo(b *testing.B) {
	logger, err := f1.NewLogger(f1.Config{})
	if err != nil {
		panic(err)
	}
	defer logger.Close()
	logger.ErrorHandler = f1.ErrorHandlerFunc(func(err error, _ []*f1.Message) error {
		panic(err)
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := logger.Post("aaaa", structmsg); err != nil {
			panic(err)
		}
	}
}

func BenchmarkStructFluentLoggerGolang(b *testing.B) {
	logger, err := f2.New(f2.Config{})
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := logger.Post("aaaa", structmsg); err != nil {
			panic(err)
		}
	}
}

func BenchmarkMapFluentLoggerGo(b *testing.B) {
	logger, err := f1.NewLogger(f1.Config{})
	if err != nil {
		panic(err)
	}
	defer logger.Close()
	logger.ErrorHandler = f1.ErrorHandlerFunc(func(err error, _ []*f1.Message) error {
		panic(err)
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := logger.Post("aaaa", mapmsg); err != nil {
			panic(err)
		}
	}
}

func BenchmarkMapFluentLoggerGolang(b *testing.B) {
	logger, err := f2.New(f2.Config{})
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := logger.Post("aaaa", mapmsg); err != nil {
			panic(err)
		}
	}
}
