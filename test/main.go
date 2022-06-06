package main

import (
	"fmt"
	"time"

	"github.com/lesismal/perf"
)

func main() {
	recorder := perf.NewRecorder()
	collector, err := perf.NewCollector(0)
	if err != nil {
		panic(err)
	}

	recorder.Warmup(100, 30000, func() error {
		time.Sleep(time.Second / 1000)
		return nil
	})

	collector.Start(true, true, true, time.Second)

	recorder.Benchmark(1000, 1000000, func() error {
		time.Sleep(time.Second / 10000)
		return nil
	})

	collector.Stop()

	fmt.Println("-------------------------")
	fmt.Println("recorder:")
	fmt.Println(recorder.String())
	fmt.Println("-------------------------")
	fmt.Println("collector:")
	fmt.Println(collector.Json())
}
