package main

import (
	"fmt"
	"go-net-benchmark/perf"
	"time"
)

func main() {
	recorder := perf.NewRecorder()
	collector, err := perf.NewCollector(0)
	if err != nil {
		panic(err)
	}

	recorder.Warmup(100, 300, func() {
		time.Sleep(time.Second / 10)
	})

	collector.Start(true, true, true, time.Second)

	recorder.Benchmark(100, 300, func() {
		time.Sleep(time.Second)
	})

	collector.Stop()

	fmt.Println("-------------------------")
	fmt.Println("recorder:")
	fmt.Println(recorder.Json())
	fmt.Println("-------------------------")
	fmt.Println("collector:")
	fmt.Println(collector.Json())
}
