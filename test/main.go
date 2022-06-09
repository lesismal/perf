package main

import (
	"fmt"
	"time"

	"github.com/lesismal/perf"
)

func main() {
	calculator := perf.NewCalculator("test")
	process, err := perf.NewProcess(0)
	if err != nil {
		panic(err)
	}

	calculator.Warmup(100, 1000, func() error {
		time.Sleep(time.Second / 1000)
		return nil
	})

	process.Start(true, true, true, time.Second)

	calculator.Benchmark(100, 2000, func() error {
		time.Sleep(time.Second / 1000)
		return nil
	}, []int{50, 60, 70, 80, 90, 95, 99, 999})

	process.Stop()

	fmt.Println("-------------------------")
	fmt.Println(calculator.String())
	fmt.Printf("TP50: %.2fms\n", float64(calculator.TPN(50))/1000000.0)
	fmt.Println("-------------------------")
	fmt.Println(process.Json())
	fmt.Println("-------------------------")
	table := perf.NewTable()
	table.SetTitle([]string{"GoNet", "TP50", "TP99", "CPU", "MEM"})
	table.AddRow([]string{"---", "---", "---", "---", "---"})
	table.AddRow([]string{"net", "3.12ms", "8.97", "1.3%", "17m"})
	table.AddRow([]string{"nbio", "12.12ms", "22.95", "1.21%", "7m"})
	table.AddRow([]string{"gnet", "13.54ms", "25.767", "1.32%", "7m"})
	fmt.Println(table.String())
}
