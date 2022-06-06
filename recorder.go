package perf

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Recorder struct {
	Ret []int64
}

func (r *Recorder) Warmup(concurrency, times int, executor func()) {
	r.benchmark(concurrency, times, func(cnt int) {
		executor()
	})
}

func (r *Recorder) Benchmark(concurrency, times int, executor func()) {
	r.benchmark(concurrency, times, func(cnt int) {
		idx := cnt - 1
		t := time.Now()
		executor()
		r.Ret[idx] = time.Since(t).Nanoseconds()
	})
}

func (r *Recorder) benchmark(concurrency, times int, executor func(cnt int)) {
	var (
		total uint64
		wg    sync.WaitGroup
	)

	r.Ret = make([]int64, times)

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
	return fmt.Sprintf("%v", r.Ret)
}

func (r *Recorder) Json() string {
	b, err := json.MarshalIndent(r.Ret, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func NewRecorder() *Recorder {
	return &Recorder{}
}
