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

type collectorResult struct {
	RetCPU []uint64                  `json:"cpu"`
	RetMEM []*process.MemoryInfoStat `json:"mem"`
	RetNET map[string][][]uint64     `json:"net"`
}

type Collector struct {
	sync.WaitGroup
	collectorResult
	proc   *process.Process
	cancel func()
}

func (c *Collector) Start(cpu, mem, net bool, interval time.Duration) {
	c.Add(1)
	defer c.Done()

	c.RetCPU = make([]uint64, 0)
	c.RetMEM = make([]*process.MemoryInfoStat, 0)
	c.RetNET = make(map[string][][]uint64)

	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	if cpu {
		c.Add(1)
		go func() {
			defer c.Done()
			ticker := time.NewTicker(interval)
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					percent, err := c.proc.CPUPercent()
					if err == nil {
						c.RetCPU = append(c.RetCPU, uint64(percent*1000))
					}
				}
			}
		}()
	}

	if mem {
		c.Add(1)
		go func() {
			defer c.Done()
			ticker := time.NewTicker(interval)
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					stat, err := c.proc.MemoryInfo()
					if err == nil {
						c.RetMEM = append(c.RetMEM, stat)
					}
				}
			}
		}()
	}

	if net {
		c.Add(1)
		go func() {
			defer c.Done()
			ticker := time.NewTicker(interval)
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					stats, err := c.proc.NetIOCounters(false)
					if err == nil {
						for _, stat := range stats {
							if c.RetNET[stat.Name] == nil {
								c.RetNET[stat.Name] = make([][]uint64, 2)
								c.RetNET[stat.Name][0] = make([]uint64, 0)
								c.RetNET[stat.Name][1] = make([]uint64, 0)
								c.RetNET[stat.Name][2] = make([]uint64, 0)
								c.RetNET[stat.Name][3] = make([]uint64, 0)
							}
							c.RetNET[stat.Name][0] = append(c.RetNET[stat.Name][0], stat.BytesRecv)
							c.RetNET[stat.Name][1] = append(c.RetNET[stat.Name][1], stat.BytesRecv)
							c.RetNET[stat.Name][2] = append(c.RetNET[stat.Name][2], stat.PacketsRecv)
							c.RetNET[stat.Name][3] = append(c.RetNET[stat.Name][3], stat.PacketsSent)
						}
					}
				}
			}
		}()
	}
}

func (c *Collector) Stop() {
	if c.cancel != nil {
		time.Sleep(time.Second / 10)
		c.cancel()
	}
	c.Wait()
}

func (c *Collector) String() string {
	return fmt.Sprintf("%v", c.collectorResult)
}

func (c *Collector) Json() string {
	b, err := json.MarshalIndent(c.collectorResult, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func NewCollector(pid int) (*Collector, error) {
	if pid == 0 {
		pid = os.Getpid()
	}
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		return nil, err
	}
	return &Collector{
		proc: proc,
	}, nil
}
