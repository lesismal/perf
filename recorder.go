package perf

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Recorder struct {
	Consumption []int64
	FailedCount int64
}

func (r *Recorder) Warmup(concurrency, times int, executor func() error) {
	r.benchmark(concurrency, times, func(cnt int) {
		executor()
	})
}

func (r *Recorder) Benchmark(concurrency, times int, executor func() error) {
	r.benchmark(concurrency, times, func(cnt int) {
		idx := cnt - 1
		t := time.Now()
		err := executor()
		r.Consumption[idx] = time.Since(t).Nanoseconds()
		if err != nil {
			atomic.AddInt64(&r.FailedCount, 1)
		}
	})
}

func (r *Recorder) benchmark(concurrency, times int, executor func(cnt int)) {
	var (
		total uint64
		wg    sync.WaitGroup
	)

	r.Consumption = make([]int64, times)

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
}

func (r *Recorder) String() string {
	var avg int64
	var sum int64
	var min = r.Consumption[0]
	var max = r.Consumption[len(r.Consumption)-1]

	sort.Slice(r.Consumption, func(i, j int) bool {
		return r.Consumption[i] < r.Consumption[j]
	})

	for _, v := range r.Consumption {
		sum += v
	}
	avg = sum / int64(len(r.Consumption))
	return fmt.Sprintf("min: %v\nmax: %v\navg: %v", min, max, avg)
}

func (r *Recorder) Json() string {
	b, err := json.MarshalIndent(r.Consumption, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func NewRecorder() *Recorder {
	return &Recorder{}
}
