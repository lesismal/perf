package perf

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

type psResult struct {
	RetCPU []float64                        `json:"cpu"`
	RetMEM []*process.MemoryInfoStat        `json:"mem"`
	RetIO  []*process.IOCountersStat        `json:"io"`
	RetNET map[string][]*net.IOCountersStat `json:"net"`
}

type PSCountOptions struct {
	CountCPU bool
	CountMEM bool
	CountIO  bool
	CountNET bool
	Interval time.Duration
}

type PSCounter struct {
	sync.WaitGroup
	psResult
	proc   *process.Process
	cancel func()
}

func (p *PSCounter) Start(opt PSCountOptions) {
	p.Add(1)
	defer p.Done()

	p.RetCPU = make([]float64, 0)
	p.RetMEM = make([]*process.MemoryInfoStat, 0)
	p.RetIO = make([]*process.IOCountersStat, 0)
	p.RetNET = make(map[string][]*net.IOCountersStat)

	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	if opt.Interval <= 0 {
		opt.Interval = time.Second
	}

	if opt.CountCPU {
		p.Add(1)
		go func() {
			defer p.Done()
			ticker := time.NewTicker(opt.Interval)
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					percent, err := p.proc.CPUPercent()
					if err == nil {
						p.RetCPU = append(p.RetCPU, percent)
					}
				}
			}
		}()
	}

	if opt.CountMEM {
		p.Add(1)
		go func() {
			defer p.Done()
			ticker := time.NewTicker(opt.Interval)
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					stat, err := p.proc.MemoryInfo()
					if err == nil {
						p.RetMEM = append(p.RetMEM, stat)
					}
				}
			}
		}()
	}

	if opt.CountIO {
		p.Add(1)
		go func() {
			defer p.Done()
			ticker := time.NewTicker(opt.Interval)
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					stat, err := p.proc.IOCounters()
					if err == nil {
						p.RetIO = append(p.RetIO, stat)
					}
				}
			}
		}()
	}

	if opt.CountNET {
		p.Add(1)
		go func() {
			defer p.Done()
			ticker := time.NewTicker(opt.Interval)
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					stats, err := p.proc.NetIOCounters(false)
					if err == nil {
						for _, stat := range stats {
							if p.RetNET[stat.Name] == nil {
								p.RetNET[stat.Name] = make([]*net.IOCountersStat, 0)
							}
							p.RetNET[stat.Name] = append(p.RetNET[stat.Name], &stat)
						}
					}
				}
			}
		}()
	}
}

func (p *PSCounter) Stop() {
	if p.cancel != nil {
		// time.Sleep(time.Second / 100)
		p.cancel()
	}
	p.Wait()
}

func (p *PSCounter) String() string {
	return fmt.Sprintf("%v", p.psResult)
}

func (p *PSCounter) Json() string {
	b, err := json.MarshalIndent(p.psResult, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func NewPSCounter(pid int) (*PSCounter, error) {
	if pid == 0 {
		pid = os.Getpid()
	}
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		return nil, err
	}
	return &PSCounter{
		proc: proc,
	}, nil
}
