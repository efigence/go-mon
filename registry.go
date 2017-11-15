package mon

import (
	"sync"
	"fmt"
	"time"
	"runtime"
)


var Registry map[string]Metric
var registryLock = &sync.RWMutex{}

func init() {
	Registry = make(map[string]Metric)
}

func Register(name string, metric Metric) (Metric, error)  {
	registryLock.Lock()
	defer registryLock.Unlock()
	if r, ok := Registry[name]; ok {
		return r, fmt.Errorf("there is metric already registered under this name")
	}
	Registry[name] = metric
	return metric, nil
}


func RegisterGcStats(t ...time.Duration) {
	interval := time.Second * 10
	if len(t) > 0 {
		interval = t[0]
	}
	go func (){
		stats := &runtime.MemStats{}
		gcCount, _ := Register(`gc.count`, NewRawCounter(0))
		gcPause, _ := Register(`gc.pause_ns`, NewRawCounterFloat(0))
		gcCpuPercentage, _ := Register(`gc.cpu_percent`, NewEWMA(time.Minute))
		heapAlloc, _ := Register(`gc.heap_alloc`, NewEWMA(time.Minute))
		heapIdle, _ := Register(`gc.heap_idle`, NewEWMA(time.Minute))
		heapInuse, _ := Register(`gc.heap_inuse`, NewEWMA(time.Minute))
		for {
			runtime.ReadMemStats(stats)
			gcCount.Update(stats.NumGC)
			gcPause.Update(float64(stats.PauseTotalNs))
			gcCpuPercentage.Update(stats.GCCPUFraction)
			heapAlloc.Update(stats.HeapAlloc)
			heapIdle.Update(stats.HeapIdle)
			heapInuse.Update(stats.HeapInuse)
			time.Sleep(interval)
		}
	} ()


}
