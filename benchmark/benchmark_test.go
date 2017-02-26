package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"testing"
	"time"

	daichirata "github.com/daichirata/fluent-logger-go"
	official "github.com/fluent/fluent-logger-golang/fluent"
)

func init() {
	if os.Getenv("DUMMY_DAEMON") == "1" {
		fmt.Println("enable dummay daemon")
		runDummyDaemon()
	}
}

func runDummyDaemon() {
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
	A string  `msg:"a"`
	B string  `msg:"b"`
	C int     `msg:"c"`
	D int     `msg:"d"`
	E string  `msg:"e"`
	F string  `msg:"f"`
	G []int   `msg:"g"`
	H int64   `msg:"h"`
	I int64   `msg:"i"`
	J int64   `msg:"j"`
	K int64   `msg:"k"`
	L int64   `msg:"l"`
	N int64   `msg:"n"`
	M int64   `msg:"m"`
	O int64   `msg:"o"`
	P int     `msg:"p"`
	Q int     `msg:"q"`
	R int64   `msg:"r"`
	S int     `msg:"s"`
	T int     `msg:"t"`
	U float64 `msg:"u"`
	V string  `msg:"v"`
	W int     `msg:"w"`
	X int     `msg:"x"`
	Y string  `msg:"y"`
	Z string  `msg:"z"`
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
	Z: "string",
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
	"Z": "string",
}

func BenchmarkStructDaichirata(b *testing.B) {
	logger, err := daichirata.NewLogger(daichirata.Config{})
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := logger.Post("debug.daichirata", structmsg); err != nil {
			panic(err)
		}
	}
}

func BenchmarkStructOfficial(b *testing.B) {
	logger, err := official.New(official.Config{})
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := logger.Post("debug.daichirata", structmsg); err != nil {
			panic(err)
		}
	}
}

func BenchmarkMapDaichirata(b *testing.B) {
	logger, err := daichirata.NewLogger(daichirata.Config{})
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := logger.Post("debug.official", mapmsg); err != nil {
			panic(err)
		}
	}
}

func BenchmarkMapOfficial(b *testing.B) {
	logger, err := official.New(official.Config{})
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := logger.Post("debug.official", mapmsg); err != nil {
			panic(err)
		}
	}
}
