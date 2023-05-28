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

func (c *Calculator) Warmup(concurrent, times int, executor func() error) {
	c.benchmark(concurrent, times, true, func(cnt int) {
		executor()
	})
}

func (c *Calculator) Benchmark(concurrent, times int, executor func() error, percents []int) {
	c.benchmark(concurrent, times, false, func(cnt int) {
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

func (c *Calculator) benchmark(concurrent, times int, warmup bool, executor func(cnt int)) {
	var (
		total uint64
		wg    sync.WaitGroup
	)

	c.Cost = make([]int64, times)

	begin := time.Now()
	for i := 0; i < concurrent; i++ {
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
	c.Avg = Avg(c.Cost)

	for _, k := range c.percents {
		c.tp[k] = c.TPN(k)
	}
}

func (c *Calculator) TPS() int64 {
	return int64(float64(c.Total) / c.Used.Seconds())
}

func (c *Calculator) TPN(percent int) int64 {
	if c.tp == nil {
		c.tp = map[int]int64{}
	}
	if v, ok := c.tp[percent]; ok {
		return v
	}
	cost := TPNFrom(c.Cost, percent, true)
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
SUCCESS  : %v, %3.2f%%
FAILED   : %v, %3.2f%%
TPS      : %v
TIME USED: %v
MIN USED : %.2fms
MAX USED : %.2fms
AVG USED : %.2fms`,
		c.Name,
		len(c.Cost),
		c.Success, float64(c.Success)/float64(len(c.Cost))*100.0,
		c.Failed, float64(c.Failed)/float64(len(c.Cost))*100.0,
		c.TPS(),
		usedStr,
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

func Min(cost []int64) int64 {
	if len(cost) == 0 {
		return 0
	}
	vMin := cost[0]
	for i := 1; i < len(cost); i++ {
		v := cost[i]
		if v < vMin {
			vMin = v
		}
	}
	return vMin
}

func Max(cost []int64) int64 {
	if len(cost) == 0 {
		return 0
	}
	vMax := cost[0]
	for i := 1; i < len(cost); i++ {
		v := cost[i]
		if v > vMax {
			vMax = v
		}
	}
	return vMax
}

func Avg(cost []int64) int64 {
	if len(cost) == 0 {
		return 0
	}
	var sum int64
	for _, v := range cost {
		sum += v
	}
	return sum / int64(len(cost))
}

func Sum(cost []int64) int64 {
	var sum int64
	for _, v := range cost {
		sum += v
	}
	return sum
}

func TPNFrom(cost []int64, percent int, sorted bool) int64 {
	if !sorted {
		sort.Slice(cost, func(i, j int) bool {
			return cost[i] < cost[j]
		})
	}
	base := 100
	shift := percent / 100
	for shift > 0 {
		base *= 10
		shift /= 10
	}
	idx := int(float64(percent) / float64(base) * float64(len(cost)))
	if idx >= len(cost) {
		idx = len(cost) - 1
	}
	return cost[idx]
}
