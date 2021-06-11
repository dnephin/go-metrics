package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/armon/go-metrics"
)

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	inm := metrics.NewInmemSink(10*time.Second, time.Minute)
	logger := stdoutLogger{}

	ctx := context.Background()
	produce := func() {
		for i := float32(0); ctx.Err() == nil; i++ {
			inm.SetGaugeWithLabels([]string{"gauge", "foo"}, 20+i, []metrics.Label{{"a", "b"}})
			inm.EmitKey([]string{"key", "foo"}, 30+i)
			inm.IncrCounterWithLabels([]string{"counter", "bar"}, 40+i, []metrics.Label{{"a", "b"}})
			inm.IncrCounterWithLabels([]string{"counter", "bar"}, 50+i, []metrics.Label{{"a", "b"}})
			inm.AddSampleWithLabels([]string{"sample", "bar"}, 60+i, []metrics.Label{{"a", "b"}})
			inm.AddSampleWithLabels([]string{"sample", "bar"}, 70+i, []metrics.Label{{"a", "b"}})
			time.Sleep(20 * time.Millisecond)
		}
	}

	go produce()
	go produce()
	go produce()

	s := &http.Server{
		Addr: "localhost:8080",
		Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			inm.Stream(req.Context(), logger, resp)
		}),
	}
	return s.ListenAndServe()
}

type stdoutLogger struct{}

func (stdoutLogger) Warn(msg string, args ...interface{}) {
	fmt.Print(msg)
	fmt.Println(args...)
}
