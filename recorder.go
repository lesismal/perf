package perf

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Recorder struct {
	Name     string
	Total    int
	Used     time.Duration
	Min      int64
	Max      int64
	Avg      int64
	Success  int64
	Failed   int64
	percents []int
	TP       map[int]int64
	Cost     []int64 `json:"-"`
}

func (r *Recorder) Warmup(concurrency, times int, executor func() error) {
	r.benchmark(concurrency, times, true, func(cnt int) {
		executor()
	})
}

func (r *Recorder) Benchmark(concurrency, times int, executor func() error) {
	r.benchmark(concurrency, times, false, func(cnt int) {
		idx := cnt - 1
		t := time.Now()
		err := executor()
		r.Cost[idx] = time.Since(t).Nanoseconds()
		atomic.AddInt64(&r.Success, 1)
		if err != nil {
			atomic.AddInt64(&r.Failed, 1)
		}
	})
}

func (r *Recorder) benchmark(concurrency, times int, warmup bool, executor func(cnt int)) {
	var (
		total uint64
		wg    sync.WaitGroup
	)

	r.Cost = make([]int64, times)

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
		r.Total = times
		r.Used = end.Sub(begin)
	}
}

func (r *Recorder) Calculate(percents []int) string {
	r.TP = map[int]int64{}
	r.percents = percents
	for i, v := range percents {
		if v < 0 {
			v = 0
		}
		if v > 100 {
			v = 100
		}
		r.TP[v] = 0
		r.percents[i] = v
	}

	sort.Slice(r.Cost, func(i, j int) bool {
		return r.Cost[i] < r.Cost[j]
	})

	r.Min = r.Cost[0]
	r.Max = r.Cost[len(r.Cost)-1]

	var sum int64
	for _, v := range r.Cost {
		sum += v
	}
	r.Avg = sum / int64(len(r.Cost))

	for _, k := range r.percents {
		base := 100
		shift := k / 100
		for shift > 0 {
			base *= 10
			shift /= 10
		}
		idx := int(float64(k) / float64(base) * float64(len(r.Cost)))
		if idx >= len(r.Cost) {
			idx = len(r.Cost) - 1
		}
		r.TP[k] = r.Cost[idx]
	}

	return r.String()
}

func (r *Recorder) String() string {
	used := r.Used.Seconds()
	usedStr := fmt.Sprintf("%.2fs", used)
	if used < 1.0 {
		used = float64(r.Used.Milliseconds())
		usedStr = fmt.Sprintf("%.2fms", used)
	}
	s := fmt.Sprintf(`NAME     : %v
BENCHMARK: %v times
TIME USED: %v
SUCCESS  : %v, %3.2f%%
FAILED   : %v, %3.2f%%
MIN      : %.2fms
MAX      : %.2fms
AVG      : %.2fms`,
		r.Name,
		len(r.Cost),
		usedStr,
		r.Success, float64(r.Success)/float64(len(r.Cost))*100.0,
		r.Failed, float64(r.Failed)/float64(len(r.Cost))*100.0,
		float64(r.Min)/1000000.0,
		float64(r.Max)/1000000.0,
		float64(r.Avg)/1000000.0)

	l := len("BENCHMARK")
	for _, k := range r.percents {
		tp := fmt.Sprintf("TP%v", k)
		for len(tp) < l {
			tp += " "
		}
		s += fmt.Sprintf("\n%v: %.2fms", tp, float64(r.TP[k])/1000000.0)
	}

	return s
}

// func (r *Recorder) Json() string {
// 	b, err := json.MarshalIndent(r.Cost, "", "  ")
// 	if err != nil {
// 		return err.Error()
// 	}
// 	return string(b)
// }

func NewRecorder(name string) *Recorder {
	return &Recorder{Name: name}
}
