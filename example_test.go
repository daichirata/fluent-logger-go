package fluent_test

import (
	"log"
	"time"

	"github.com/daichirata/fluent-logger-go"
)

func ExampleNewLogger() {
	_, err := fluent.NewLogger(fluent.Config{})
	if err != nil {
		// TODO: Handle error.
	}
	// See the other examples to learn how to use the Logger.
}

func ExampleNewLogger_errorhandler() {
	logger, err := fluent.NewLogger(fluent.Config{})
	if err != nil {
		// TODO: Handle error.
	}

	// Logging error.
	logger.ErrorHandler = fluent.ErrorHandlerFunc(func(err error, _ []byte) error {
		log.Println(err)
		return err
	})

	// Fallback logger.
	fallback, err := fluent.NewLogger(fluent.Config{})
	if err != nil {
		// TODO: Handle error.
	}
	logger.ErrorHandler = fluent.NewFallbackHandler(fallback)
}

func ExampleLogger_Post_map() {
	logger, err := fluent.NewLogger(fluent.Config{})
	if err != nil {
		// TODO: Handle error.
	}

	err = logger.Post("tag-name", map[string]string{
		"foo": "bar",
	})
	if err != nil {
		// TODO: Handle error.
	}
}

func ExampleLogger_Post_struct() {
	logger, err := fluent.NewLogger(fluent.Config{})
	if err != nil {
		// TODO: Handle error.
	}

	data := struct {
		Foo string `msgpack:"foo"`
		Bar string `msgpack:"bar"`
	}{
		Foo: "foo",
		Bar: "bar",
	}
	err = logger.Post("tag-name", data)
	if err != nil {
		// TODO: Handle error.
	}
}

func ExampleLogger_PostWithTime() {
	logger, err := fluent.NewLogger(fluent.Config{})
	if err != nil {
		// TODO: Handle error.
	}

	data := struct {
		Foo string `msgpack:"foo"`
		Bar string `msgpack:"bar"`
	}{
		Foo: "foo",
		Bar: "bar",
	}
	// Post a message with specified time
	err = logger.PostWithTime("tag-name", time.Now().Add(-5*time.Second), data)
	if err != nil {
		// TODO: Handle error.
	}
}

func ExampleLogger_Close() {
	logger, err := fluent.NewLogger(fluent.Config{})
	if err != nil {
		// TODO: Handle error.
	}

	for i := 0; i < 10; i++ {
		err = logger.Post("tag-name", map[string]string{
			"foo": "bar",
		})
		if err != nil {
			// TODO: Handle error.
		}
	}

	// Close will block until all messages has been sent.
	err = logger.Close()
	if err != nil {
		// TODO: Handle error.
	}
}
