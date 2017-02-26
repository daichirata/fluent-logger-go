package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/daichirata/fluent-logger-go"
)

var (
	tag  = flag.String("tag", "", "")
	addr = flag.String("addr", "", "")
)

func main() {
	flag.Parse()
	if *tag == "" {
		log.Fatal("Flag --tag is required")
	}

	logger, err := fluent.NewLogger(fluent.Config{
		Address: *addr,
		// ErrorHandler: fluent.NewFallbackHandler(os.Stdout),
	})
	if err != nil {
		log.Fatal(err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)
	go func() {
		select {
		case <-sigCh:
			logger.Close()
			os.Exit(0)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		logger.Post(*tag, map[string]string{"message": scanner.Text()})
	}
}
