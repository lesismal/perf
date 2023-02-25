package main

import (
	"fmt"
	"time"

	"github.com/lesismal/perf"
)

func main() {
	psCounter, err := perf.NewPSCounter(0)
	if err != nil {
		panic(err)
	}

	calculator := perf.NewCalculator("test")
	calculator.Warmup(100, 1000, func() error {
		time.Sleep(time.Second / 1000)
		return nil
	})

	psCounter.Start(perf.PSCountOptions{
		CountCPU: true,
		CountMEM: true,
		CountIO:  true,
		CountNET: true,
		Interval: time.Second,
	})
	calculator.Benchmark(10000, 1000000, func() error {
		time.Sleep(time.Second / 50)
		return nil
	}, []int{50, 60, 70, 80, 90, 95, 99, 999})
	psCounter.Stop()

	fmt.Println("-------------------------")
	fmt.Println(calculator.String())
	fmt.Printf("TP50: %.2fms\n", float64(calculator.TPN(50))/1000000.0)
	fmt.Println("-------------------------")
	fmt.Println(psCounter.Json())
	fmt.Println("-------------------------")
	fmt.Println("CPUMin:", psCounter.CPUMin())
	fmt.Println("CPUMax:", psCounter.CPUMax())
	fmt.Println("CPUAvg:", psCounter.CPUAvg())
	fmt.Println("-------------------------")
	fmt.Println("MEMRSSMin:", psCounter.MEMRSSMin())
	fmt.Println("MEMRSSMax :", psCounter.MEMRSSMax())
	fmt.Println("MEMRSSAvg :", psCounter.MEMRSSAvg())
	fmt.Println("-------------------------")
	fmt.Println("MEMVMSMin:", psCounter.MEMVMSMin())
	fmt.Println("MEMVMSMax :", psCounter.MEMVMSMax())
	fmt.Println("MEMVMSAvg :", psCounter.MEMVMSAvg())
	fmt.Println("-------------------------")
	fmt.Println("IOReadCountMin:", psCounter.IOReadCountMin())
	fmt.Println("IOReadCountMax :", psCounter.IOReadCountMax())
	fmt.Println("IOReadCountAvg :", psCounter.IOReadCountAvg())
	fmt.Println("-------------------------")
	fmt.Println("IOReadBytesMin:", psCounter.IOReadBytesMin())
	fmt.Println("IOReadBytesMax :", psCounter.IOReadBytesMax())
	fmt.Println("IOReadBytesAvg :", psCounter.IOReadBytesAvg())
	fmt.Println("-------------------------")
	fmt.Println("IOWriteCountMin:", psCounter.IOWriteCountMin())
	fmt.Println("IOWriteCountMax :", psCounter.IOWriteCountMax())
	fmt.Println("IOWriteCountAvg :", psCounter.IOWriteCountAvg())
	fmt.Println("-------------------------")
	fmt.Println("IOWriteBytesMin:", psCounter.IOWriteBytesMin())
	fmt.Println("IOWriteBytesMax :", psCounter.IOWriteBytesMax())
	fmt.Println("IOWriteBytesAvg :", psCounter.IOWriteBytesAvg())
	fmt.Println("-------------------------")
	table := perf.NewTable()
	table.SetTitle([]string{"Frameworks", "TP50", "TP99", "CPU", "MEM"})
	table.AddRow([]string{"net", "3.12ms", "8.97", "1.3%", "17m"})
	table.AddRow([]string{"nbio", "12.12ms", "22.95", "1.21%", "7m"})
	table.AddRow([]string{"gnet", "13.54ms", "25.767", "1.32%", "7m"})
	fmt.Println(table.Markdown())
}
