package mon

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var GlobalRegistry *Registry

type Registry struct {
	Metrics  map[string]Metric `json:"metrics"`
	Instance string            `json:"instance"`
	sync.Mutex
}

func (r *Registry) GetMetric(name string) (Metric, error) {
	if r, ok := r.Metrics[name]; ok {
		return r, nil
	} else {
		return nil, &ErrMetricNotFound{Metric: name}
	}
}

func (r *Registry) SetInstance(name string) {
	r.Lock()
	r.Instance = name
	r.Unlock()
}

// Register() a given metric or return error if name is already used
func (r *Registry) Register(name string, metric Metric) (Metric, error) {
	r.Lock()
	defer r.Unlock()
	if r, ok := r.Metrics[name]; ok {
		return r, &ErrMetricAlreadyRegistered{Metric: name}
	}
	r.Metrics[name] = metric
	return metric, nil
}

// RegisterOrGet() registers a given metric or resturns already existing one if it is of same type
// it will err out if type does not match but it does not compare rest of the parameters of the metric so do not use it if you are not 100% sure

func (r *Registry) RegisterOrGet(name string, metric Metric) (Metric, error) {
	r.Lock()
	defer r.Unlock()
	if r, ok := r.Metrics[name]; ok {
		if r.Type() == metric.Type() {
			return r, nil
		}
		return r, &ErrMetricAlreadyRegisteredWrongType{
			Metric:        name,
			OldMetricType: r.Type(),
			NewMetricType: metric.Type(),
		}
	}
	r.Metrics[name] = metric
	return metric, nil
}

// MustRegister() does same as Register() except it panic()s if metric already exists.
// It is mostly intended to be used for top of the package, package-scoped metrics like
//  var request_rate =  mon.GlobalRegistry.Register("backend.mysql.qps",mon.NewEWMARate(time.Duration(time.Minute)))
//
func (r *Registry) MustRegister(name string, metric Metric) Metric {
	r.Lock()
	defer r.Unlock()
	if r, ok := r.Metrics[name]; ok {
		panic(fmt.Sprintf("Metric is already registered: %s : %+v", name, r))
	}
	r.Metrics[name] = metric
	return metric
}

func init() {
	_, name := filepath.Split(os.Args[0])
	GlobalRegistry = &Registry{
		Instance: name,
		Metrics:  make(map[string]Metric),
	}
}

func RegisterGcStats(t ...time.Duration) {
	interval := time.Second * 10
	if len(t) > 0 {
		interval = t[0]
	}
	// make sure metrics actually exist in registry at the moment of exit
	stats := &runtime.MemStats{}
	gcCount, _ := GlobalRegistry.Register(`gc.count`, NewRawCounter())
	gcPause, _ := GlobalRegistry.Register(`gc.pause`, NewRawCounterFloat("duration"))
	gcCPUPercentage, _ := GlobalRegistry.Register(`gc.cpu`, NewEWMA(time.Minute, "percent"))
	heapAlloc, _ := GlobalRegistry.Register(`gc.heap_alloc`, NewEWMA(time.Minute, "bytes"))
	heapIdle, _ := GlobalRegistry.Register(`gc.heap_idle`, NewEWMA(time.Minute, "bytes"))
	heapInuse, _ := GlobalRegistry.Register(`gc.heap_inuse`, NewEWMA(time.Minute, "bytes"))
	go func() {
		for {
			runtime.ReadMemStats(stats)
			gcCount.Update(stats.NumGC)
			gcPause.Update(float64(stats.PauseTotalNs) / 1000000000)
			gcCPUPercentage.Update(stats.GCCPUFraction)
			heapAlloc.Update(stats.HeapAlloc)
			heapIdle.Update(stats.HeapIdle)
			heapInuse.Update(stats.HeapInuse)
			time.Sleep(interval)
		}
	}()
}
