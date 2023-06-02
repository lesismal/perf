package perf

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Calculator struct {
	Name         string
	Total        int
	Used         time.Duration
	Min          int64
	Max          int64
	Avg          int64
	Success      int64
	Failed       int64
	FailedErrors map[string]int
	Cost         []int64 `json:"-"`
	tp           map[int]int64
	percents     []int
	result       string
}

func (c *Calculator) Warmup(concurrent, times int, executor func() error) {
	c.benchmark(concurrent, times, func(cnt int) {
		executor()
	})
}

func (c *Calculator) Benchmark(concurrent, times int, executor func() error, percents []int) {
	begin := time.Now()
	mux := sync.Mutex{}
	c.Cost = make([]int64, times)
	c.FailedErrors = map[string]int{}
	c.benchmark(concurrent, times, func(cnt int) {
		idx := cnt - 1
		t := time.Now()
		err := executor()
		if err != nil {
			atomic.AddInt64(&c.Failed, 1)
			c.Cost[idx] = -1
			mux.Lock()
			errStr := err.Error()
			errCnt := c.FailedErrors[errStr]
			c.FailedErrors[errStr] = errCnt + 1
			mux.Unlock()
		} else {
			c.Cost[idx] = time.Since(t).Nanoseconds()
			atomic.AddInt64(&c.Success, 1)
		}
	})
	end := time.Now()
	c.Total = times
	c.Used = end.Sub(begin)
	c.calculate(percents)
}

func (c *Calculator) benchmark(concurrent, times int, executor func(cnt int)) {
	var (
		total uint64
		wg    sync.WaitGroup
	)

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

	var min, max int64
	sort.Slice(c.Cost, func(i, j int) bool {
		if (min == 0 && c.Cost[i] > 0) || c.Cost[i] < min {
			min = c.Cost[i]
		}
		if (min == 0 && c.Cost[j] > 0) || c.Cost[j] < min {
			min = c.Cost[j]
		}
		if c.Cost[i] > max {
			max = c.Cost[i]
		}
		if c.Cost[j] > max {
			max = c.Cost[i]
		}
		if c.Cost[i] < 0 {
			return false
		}
		return c.Cost[i] < c.Cost[j]
	})

	c.Min = min
	c.Max = max

	var sum int64
	for _, v := range c.Cost {
		if v > 0 {
			sum += v
		}
	}
	c.Avg = Avg(c.Cost)

	for _, k := range c.percents {
		c.tp[k] = c.TPN(k)
	}
}

func (c *Calculator) TPS() int64 {
	return int64(float64(c.Success) / c.Used.Seconds())
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

	s := fmt.Sprintf(`TOTAL    : %v times
SUCCESS  : %v, %3.2f%%
FAILED   : %v, %3.2f%%
TPS      : %v
TIME USED: %v
MIN USED : %v
AVG USED : %v
MAX USED : %v`,
		c.Total,
		c.Success, float64(c.Success)/float64(len(c.Cost))*100.0,
		c.Failed, float64(c.Failed)/float64(len(c.Cost))*100.0,
		c.TPS(),
		I2TimeString(int64(c.Used)),
		I2TimeString(c.Min),
		I2TimeString(c.Avg),
		I2TimeString(c.Max))

	l := len("BENCHMARK")
	for _, k := range c.percents {
		tp := fmt.Sprintf("TP%v", k)
		for len(tp) < l {
			tp += " "
		}
		s += fmt.Sprintf("\n%v: %v", tp, I2TimeString(c.tp[k]))
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
		if v > 0 {
			if vMin <= 0 || v < vMin {
				vMin = v
			}
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
	var n int64
	var sum int64
	for _, v := range cost {
		if v > 0 {
			sum += v
			n++
		}
	}
	if n == 0 {
		return -1
	}
	return sum / n
}

func Sum(cost []int64) int64 {
	var sum int64
	for _, v := range cost {
		if v > 0 {
			sum += v
		}
	}
	return sum
}

func TPNFrom(cost []int64, percent int, args ...interface{}) int64 {
	base := 100
	shift := percent / 100
	for shift > 0 {
		base *= 10
		shift /= 10
	}

	return TPNFromBase(cost, percent, base, args...)
}

func TPNFromBase(cost []int64, percent, base int, args ...interface{}) int64 {
	sorted := false
	if len(args) > 0 {
		v, ok := args[0].(bool)
		if ok {
			sorted = v
		}
	}

	if !sorted {
		sort.Slice(cost, func(i, j int) bool {
			if cost[i] < 0 {
				return false
			}
			return cost[i] < cost[j]
		})
	}

	for i, v := range cost {
		if v < 0 {
			cost = cost[:i]
			break
		}
	}

	idx := int(float64(percent) / float64(base) * float64(len(cost)))
	if idx >= len(cost) {
		idx = len(cost) - 1
	}

	return cost[idx]
}

func I2TimeString(i int64) string {
	used := float64(i) / float64(1e9)
	usedStr := fmt.Sprintf("%.2fs", used)
	if i/1e9 < 1 {
		used = float64(i) / float64(1e6)
		usedStr = fmt.Sprintf("%.2fms", used)
		if i/1e6 < 1 {
			used = float64(i) / float64(1e3)
			usedStr = fmt.Sprintf("%.2fus", used)
			if i/1e3 < 1 {
				usedStr = fmt.Sprintf("%dns", i)
			}
		}
	}
	return usedStr
}

func I2MemString(i uint64) string {
	used := float64(i) / float64(1e9)
	usedStr := fmt.Sprintf("%.2fG", used)
	if used < 1.0 {
		used = float64(i) / float64(1e6)
		usedStr = fmt.Sprintf("%.2fM", used)
		if used < 1.0 {
			used = float64(i) / float64(1e3)
			usedStr = fmt.Sprintf("%.2fK", used)
		}
	}
	return usedStr
}
