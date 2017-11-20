package mon

import (
	"sync"
	"time"
	"runtime"
	"os"
	"path/filepath"
)


var GlobalRegistry  *Registry

type Registry struct {
	Metrics map[string]Metric `json:"metrics"`
	Instance string `json:"instance"`
	sync.Mutex
}

func (r *Registry)GetMetric(name string) (Metric, error) {
	if r, ok := r.Metrics[name]; ok {
		return r, nil
	} else {
		return nil, &ErrMetricNotFound{Metric: name}
	}
}

func (r *Registry)SetInstance(name string)  {
	r.Lock()
	r.Instance = name
	r.Unlock()
}

func (r *Registry)Register(name string, metric Metric) (Metric, error)  {
	r.Lock()
	defer r.Unlock()
	if r, ok := r.Metrics[name]; ok {
		return r, &ErrMetricAlreadyRegistered{Metric: name}
	}
	r.Metrics[name] = metric
	return metric, nil
}





func init() {
	_, name := filepath.Split(os.Args[0])
	GlobalRegistry = &Registry{
		Instance: name,
		Metrics: make(map[string]Metric),
	}
}


func RegisterGcStats(t ...time.Duration) {
	interval := time.Second * 10
	if len(t) > 0 {
		interval = t[0]
	}
	// make sure metrics actually exist in registry at the moment of exit
	stats := &runtime.MemStats{}
	gcCount, _ := GlobalRegistry.Register(`gc.count`, NewRawCounter(0))
	gcPause, _ := GlobalRegistry.Register(`gc.pause_ns`, NewRawCounterFloat(0))
	gcCPUPercentage, _ := GlobalRegistry.Register(`gc.cpu_percent`, NewEWMA(time.Minute))
	heapAlloc, _ := GlobalRegistry.Register(`gc.heap_alloc`, NewEWMA(time.Minute))
	heapIdle, _ := GlobalRegistry.Register(`gc.heap_idle`, NewEWMA(time.Minute))
	heapInuse, _ := GlobalRegistry.Register(`gc.heap_inuse`, NewEWMA(time.Minute))
	go func (){
		for {
			runtime.ReadMemStats(stats)
			gcCount.Update(stats.NumGC)
			gcPause.Update(float64(stats.PauseTotalNs))
			gcCPUPercentage.Update(stats.GCCPUFraction)
			heapAlloc.Update(stats.HeapAlloc)
			heapIdle.Update(stats.HeapIdle)
			heapInuse.Update(stats.HeapInuse)
			time.Sleep(interval)
		}
	} ()
}
