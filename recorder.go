package perf

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Calculator struct {
	Name     string
	Total    int
	Used     time.Duration
	Min      int64
	Max      int64
	Avg      int64
	Success  int64
	Failed   int64
	Cost     []int64 `json:"-"`
	tp       map[int]int64
	percents []int
	result   string
}

func (c *Calculator) Warmup(concurrency, times int, executor func() error) {
	c.benchmark(concurrency, times, true, func(cnt int) {
		executor()
	})
}

func (c *Calculator) Benchmark(concurrency, times int, executor func() error, percents []int) {
	c.benchmark(concurrency, times, false, func(cnt int) {
		idx := cnt - 1
		t := time.Now()
		err := executor()
		c.Cost[idx] = time.Since(t).Nanoseconds()
		atomic.AddInt64(&c.Success, 1)
		if err != nil {
			atomic.AddInt64(&c.Failed, 1)
		}
	})
	c.calculate(percents)
}

func (c *Calculator) benchmark(concurrency, times int, warmup bool, executor func(cnt int)) {
	var (
		total uint64
		wg    sync.WaitGroup
	)

	c.Cost = make([]int64, times)

	begin := time.Now()
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				cnt := int(atomic.AddUint64(&total, 1))
				if cnt > times {
					break
				}
				executor(cnt)
			}
		}()
	}

	wg.Wait()
	end := time.Now()

	if !warmup {
		c.Total = times
		c.Used = end.Sub(begin)
	}
}

func (c *Calculator) calculate(percents []int) {
	c.tp = map[int]int64{}
	c.percents = percents
	for i, v := range percents {
		if v < 0 {
			v = 0
		}
		if v > 100 {
			v = 100
		}
		c.percents[i] = v
	}

	sort.Slice(c.Cost, func(i, j int) bool {
		return c.Cost[i] < c.Cost[j]
	})

	c.Min = c.Cost[0]
	c.Max = c.Cost[len(c.Cost)-1]

	var sum int64
	for _, v := range c.Cost {
		sum += v
	}
	c.Avg = sum / int64(len(c.Cost))

	for _, k := range c.percents {
		base := 100
		shift := k / 100
		for shift > 0 {
			base *= 10
			shift /= 10
		}
		idx := int(float64(k) / float64(base) * float64(len(c.Cost)))
		if idx >= len(c.Cost) {
			idx = len(c.Cost) - 1
		}
		c.tp[k] = c.TPN(k)
	}
}

func (c *Calculator) TPN(percent int) int64 {
	if c.tp == nil {
		c.tp = map[int]int64{}
	}
	if v, ok := c.tp[percent]; ok {
		return v
	}
	base := 100
	shift := percent / 100
	for shift > 0 {
		base *= 10
		shift /= 10
	}
	idx := int(float64(percent) / float64(base) * float64(len(c.Cost)))
	if idx >= len(c.Cost) {
		idx = len(c.Cost) - 1
	}
	cost := c.Cost[idx]
	c.tp[percent] = cost
	return cost
}

func (c *Calculator) String() string {
	if c.result != "" {
		return c.result
	}
	used := c.Used.Seconds()
	usedStr := fmt.Sprintf("%.2fs", used)
	if used < 1.0 {
		used = float64(c.Used.Milliseconds())
		usedStr = fmt.Sprintf("%.2fms", used)
	}
	s := fmt.Sprintf(`NAME     : %v
BENCHMARK: %v times
TIME USED: %v
SUCCESS  : %v, %3.2f%%
FAILED   : %v, %3.2f%%
TPS MIN  : %.2fms
TPS MAX  : %.2fms
TPS AVG  : %.2fms`,
		c.Name,
		len(c.Cost),
		usedStr,
		c.Success, float64(c.Success)/float64(len(c.Cost))*100.0,
		c.Failed, float64(c.Failed)/float64(len(c.Cost))*100.0,
		float64(c.Min)/1000000.0,
		float64(c.Max)/1000000.0,
		float64(c.Avg)/1000000.0)

	l := len("BENCHMARK")
	for _, k := range c.percents {
		tp := fmt.Sprintf("TP%v", k)
		for len(tp) < l {
			tp += " "
		}
		s += fmt.Sprintf("\n%v: %.2fms", tp, float64(c.tp[k])/1000000.0)
	}

	c.result = s

	return s
}

// func (c *Calculator) Json() string {
// 	b, err := json.MarshalIndent(c .Cost, "", "  ")
// 	if err != nil {
// 		return erc .Error()
// 	}
// 	return string(b)
// }

func NewCalculator(name string) *Calculator {
	return &Calculator{Name: name}
}
