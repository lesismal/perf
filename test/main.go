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

	recorder.Warmup(100, 1000, func() error {
		time.Sleep(time.Second / 1000)
		return nil
	})

	collector.Start(true, true, true, time.Second)

	recorder.Benchmark(100, 2000, func() error {
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
	fmt.Println("-------------------------")
	table := perf.NewTable()
	table.SetTitle([]string{"GoNet", "TP50", "TP99", "CPU", "MEM"})
	table.AddRow([]string{"---", "---", "---", "---", "---"})
	table.AddRow([]string{"net", "3.12ms", "8.97", "1.3%", "17m"})
	table.AddRow([]string{"nbio", "12.12ms", "22.95", "1.21%", "7m"})
	table.AddRow([]string{"gnet", "13.54ms", "25.767", "1.32%", "7m"})
	fmt.Println(table.String())
}
