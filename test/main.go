package main

import (
	"fmt"
	"time"

	"github.com/lesismal/perf"
)

func main() {
	recorder := perf.NewRecorder("test")
	collector, err := perf.NewCollector(0)
	if err != nil {
		panic(err)
	}

	recorder.Warmup(100, 30000, func() error {
		time.Sleep(time.Second / 1000)
		return nil
	})

	collector.Start(true, true, true, time.Second)

	recorder.Benchmark(100, 20000, func() error {
		time.Sleep(time.Second / 1000)
		return nil
	})

	collector.Stop()

	fmt.Println("-------------------------")
	recorder.Calculate([]int{50, 60, 70, 80, 90, 95, 99, 999})
	fmt.Println(recorder.String())
	fmt.Println("-------------------------")
	recorder.Calculate([]int{50, 75, 90, 95, 999})
	fmt.Println(recorder.String())
	fmt.Println("-------------------------")
	fmt.Println(collector.Json())
}
