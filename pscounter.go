package perf

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
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

func (p *PSCounter) CPUMin() float64 {
	var ret float64 = math.MaxFloat64
	for _, v := range p.RetCPU {
		if v < ret {
			ret = v
		}
	}
	return ret
}

func (p *PSCounter) CPUMax() float64 {
	var ret float64 = 0.0
	for _, v := range p.RetCPU {
		if v > ret {
			ret = v
		}
	}
	return ret
}

func (p *PSCounter) CPUAvg() float64 {
	if len(p.RetCPU) == 0 {
		return 0.0
	}
	var ret float64
	for _, v := range p.RetCPU {
		ret += v
	}
	return ret / float64(len(p.RetCPU))
}

func (p *PSCounter) MEMRSSMin() uint64 {
	var ret uint64 = math.MaxUint64
	for _, v := range p.RetMEM {
		if v.RSS < ret {
			ret = v.RSS
		}
	}
	return ret
}

func (p *PSCounter) MEMRSSMax() uint64 {
	var ret uint64 = 0
	for _, v := range p.RetMEM {
		if v.RSS > ret {
			ret = v.RSS
		}
	}
	return ret
}

func (p *PSCounter) MEMRSSAvg() uint64 {
	if len(p.RetMEM) == 0 {
		return 0
	}
	var ret uint64
	for _, v := range p.RetMEM {
		ret += v.RSS
	}
	return ret / uint64(len(p.RetMEM))
}

func (p *PSCounter) MEMVMSMin() uint64 {
	var ret uint64 = math.MaxUint64
	for _, v := range p.RetMEM {
		if v.VMS < ret {
			ret = v.VMS
		}
	}
	return ret
}

func (p *PSCounter) MEMVMSMax() uint64 {
	var ret uint64 = 0
	for _, v := range p.RetMEM {
		if v.VMS > ret {
			ret = v.VMS
		}
	}
	return ret
}

func (p *PSCounter) MEMVMSAvg() uint64 {
	if len(p.RetMEM) == 0 {
		return 0
	}
	var ret uint64
	for _, v := range p.RetMEM {
		ret += v.VMS
	}
	return ret / uint64(len(p.RetMEM))
}

func (p *PSCounter) IOReadCountMin() uint64 {
	var ret uint64 = math.MaxUint64
	for _, v := range p.RetIO {
		if v.ReadCount < ret {
			ret = v.ReadCount
		}
	}
	return ret
}

func (p *PSCounter) IOReadCountMax() uint64 {
	var ret uint64 = 0
	for _, v := range p.RetIO {
		if v.ReadCount > ret {
			ret = v.ReadCount
		}
	}
	return ret
}

func (p *PSCounter) IOReadCountAvg() uint64 {
	if len(p.RetIO) == 0 {
		return 0
	}
	var ret uint64
	for _, v := range p.RetIO {
		ret += v.ReadCount
	}
	return ret / uint64(len(p.RetIO))
}

func (p *PSCounter) IOReadBytesMin() uint64 {
	var ret uint64 = math.MaxUint64
	for _, v := range p.RetIO {
		if v.ReadBytes < ret {
			ret = v.ReadBytes
		}
	}
	return ret
}

func (p *PSCounter) IOReadBytesMax() uint64 {
	var ret uint64 = 0
	for _, v := range p.RetIO {
		if v.ReadBytes > ret {
			ret = v.ReadBytes
		}
	}
	return ret
}

func (p *PSCounter) IOReadBytesAvg() uint64 {
	if len(p.RetIO) == 0 {
		return 0
	}
	var ret uint64
	for _, v := range p.RetIO {
		ret += v.ReadBytes
	}
	return ret / uint64(len(p.RetIO))
}

func (p *PSCounter) IOWriteCountMin() uint64 {
	var ret uint64 = math.MaxUint64
	for _, v := range p.RetIO {
		if v.WriteCount < ret {
			ret = v.WriteCount
		}
	}
	return ret
}

func (p *PSCounter) IOWriteCountMax() uint64 {
	var ret uint64 = 0
	for _, v := range p.RetIO {
		if v.WriteCount > ret {
			ret = v.WriteCount
		}
	}
	return ret
}

func (p *PSCounter) IOWriteCountAvg() uint64 {
	if len(p.RetIO) == 0 {
		return 0
	}
	var ret uint64
	for _, v := range p.RetIO {
		ret += v.WriteCount
	}
	return ret / uint64(len(p.RetIO))
}

func (p *PSCounter) IOWriteBytesMin() uint64 {
	var ret uint64 = math.MaxUint64
	for _, v := range p.RetIO {
		if v.WriteBytes < ret {
			ret = v.WriteBytes
		}
	}
	return ret
}

func (p *PSCounter) IOWriteBytesMax() uint64 {
	var ret uint64 = 0
	for _, v := range p.RetIO {
		if v.WriteBytes > ret {
			ret = v.WriteBytes
		}
	}
	return ret
}

func (p *PSCounter) IOWriteBytesAvg() uint64 {
	if len(p.RetIO) == 0 {
		return 0
	}
	var ret uint64
	for _, v := range p.RetIO {
		ret += v.WriteBytes
	}
	return ret / uint64(len(p.RetIO))
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
