package perf

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/shirou/gopsutil/process"
)

type psResult struct {
	RetCPU []uint64                  `json:"cpu"`
	RetMEM []*process.MemoryInfoStat `json:"mem"`
	RetNET map[string][][]uint64     `json:"net"`
}

type Process struct {
	sync.WaitGroup
	psResult
	proc   *process.Process
	cancel func()
}

func (p *Process) Start(cpu, mem, net bool, interval time.Duration) {
	p.Add(1)
	defer p.Done()

	p.RetCPU = make([]uint64, 0)
	p.RetMEM = make([]*process.MemoryInfoStat, 0)
	p.RetNET = make(map[string][][]uint64)

	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	if cpu {
		p.Add(1)
		go func() {
			defer p.Done()
			ticker := time.NewTicker(interval)
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					percent, err := p.proc.CPUPercent()
					if err == nil {
						p.RetCPU = append(p.RetCPU, uint64(percent*1000))
					}
				}
			}
		}()
	}

	if mem {
		p.Add(1)
		go func() {
			defer p.Done()
			ticker := time.NewTicker(interval)
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

	if net {
		p.Add(1)
		go func() {
			defer p.Done()
			ticker := time.NewTicker(interval)
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					stats, err := p.proc.NetIOCounters(false)
					if err == nil {
						for _, stat := range stats {
							if p.RetNET[stat.Name] == nil {
								p.RetNET[stat.Name] = make([][]uint64, 2)
								p.RetNET[stat.Name][0] = make([]uint64, 0)
								p.RetNET[stat.Name][1] = make([]uint64, 0)
								p.RetNET[stat.Name][2] = make([]uint64, 0)
								p.RetNET[stat.Name][3] = make([]uint64, 0)
							}
							p.RetNET[stat.Name][0] = append(p.RetNET[stat.Name][0], stat.BytesRecv)
							p.RetNET[stat.Name][1] = append(p.RetNET[stat.Name][1], stat.BytesRecv)
							p.RetNET[stat.Name][2] = append(p.RetNET[stat.Name][2], stat.PacketsRecv)
							p.RetNET[stat.Name][3] = append(p.RetNET[stat.Name][3], stat.PacketsSent)
						}
					}
				}
			}
		}()
	}
}

func (p *Process) Stop() {
	if p.cancel != nil {
		time.Sleep(time.Second / 10)
		p.cancel()
	}
	p.Wait()
}

func (p *Process) String() string {
	return fmt.Sprintf("%v", p.psResult)
}

func (p *Process) Json() string {
	b, err := json.MarshalIndent(p.psResult, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func NewProcess(pid int) (*Process, error) {
	if pid == 0 {
		pid = os.Getpid()
	}
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		return nil, err
	}
	return &Process{
		proc: proc,
	}, nil
}
