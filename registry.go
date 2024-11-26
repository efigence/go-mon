package mon

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

type Registry struct {
	Metrics  map[string]Metric `json:"metrics"`
	Instance string            `json:"instance"`
	Interval float64           `json:"interval"`
	FQDN     string            `json:"fqdn"`
	Ts       time.Time         `json:"ts,omitempty"`
	sync.Mutex
}

func NewRegistry(fqdn string, instance string, interval float64) (*Registry, error) {
	return &Registry{
		FQDN:     fqdn,
		Instance: instance,
		Interval: interval,
		Metrics:  make(map[string]Metric),
	}, nil
}

func (r *Registry) GetMetric(name string) (Metric, error) {
	if r, ok := r.Metrics[name]; ok {
		return r, nil
	} else {
		return nil, &ErrMetricNotFound{Metric: name}
	}
}

// Returns a shallow copy of registry with current timestamp.
// Should be used as source for any serializer
func (r *Registry) GetRegistry() *Registry {
	r.Lock()
	clone := Registry{
		FQDN:     r.FQDN,
		Instance: r.Instance,
		Interval: r.Interval,
		Metrics:  make(map[string]Metric),
	}
	for k, v := range r.Metrics {
		clone.Metrics[k] = v
	}
	clone.Ts = time.Now()
	r.Unlock()
	return &clone
}

// Set instance name returned by registry during marshalling
func (r *Registry) SetInstance(name string) {
	r.Lock()
	r.Instance = name
	r.Unlock()
}

// Set FQDN returned by registry during marshalling
func (r *Registry) SetFQDN(name string) {
	r.Lock()
	r.FQDN = name
	r.Unlock()
}
func (r *Registry) SetInterval(interval float64) {
	r.Lock()
	r.Interval = interval
	r.Unlock()
}

// update timestamp. Should be called before read if timestamp is desirable in output
func (r *Registry) UpdateTs() {
	// note that in this implementation metrics are wholly independent on eachother so
	// ts should always be same as time of update
	r.Lock()
	r.Ts = time.Now()
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
//
//	var request_rate =  mon.GlobalRegistry.Register("backend.mysql.qps",mon.NewEWMARate(time.Duration(time.Minute)))
func (r *Registry) MustRegister(name string, metric Metric) Metric {
	r.Lock()
	defer r.Unlock()
	if r, ok := r.Metrics[name]; ok {
		panic(fmt.Sprintf("Metric is already registered: %s : %+v", name, r))
	}
	r.Metrics[name] = metric
	return metric
}

// Register GC and memory stats under 'gc.'
// will noop if called more than one
//
// Probing occurs 3x the interval and some stats (like memory

var globalGcStatsRegistered bool

// GcStats configuration. Interval is time between probes, average turns on EWMA on most stats with 5x interval as half-life
type GcStatsConfig struct {
	Interval time.Duration
	Average  bool
}

func RegisterGcStats(c ...GcStatsConfig) {
	if globalGcStatsRegistered {
		return
	}
	interval := time.Second * 10
	EWMAHalfLife := interval * 5
	average := false

	if len(c) > 0 {
		if c[0].Interval > 0 {
			interval = c[0].Interval
		}
		average = c[0].Average
	}

	NewGaugeFunc := func(unit ...string) Metric {
		return NewRawGauge(unit...)
	}
	if average {
		NewGaugeFunc = func(unit ...string) Metric {
			return NewEWMA(EWMAHalfLife, unit...)
		}
	}
	gcCount := GlobalRegistry.MustRegister(`gc.count`, NewRawCounter())
	gcPause := GlobalRegistry.MustRegister(`gc.pause`, NewRawCounterFloat("ns"))
	gcCPUPercentage := GlobalRegistry.MustRegister(`gc.cpu`, NewRawGauge("percent"))
	mallocCount := GlobalRegistry.MustRegister(`gc.malloc`, NewRawCounter())
	freeCount := GlobalRegistry.MustRegister(`gc.free`, NewRawCounter())
	heapAlloc := GlobalRegistry.MustRegister(`gc.heap_alloc`, NewGaugeFunc("bytes"))
	heapIdle := GlobalRegistry.MustRegister(`gc.heap_idle`, NewGaugeFunc("bytes"))
	heapInuse := GlobalRegistry.MustRegister(`gc.heap_inuse`, NewGaugeFunc("bytes"))
	heapObj := GlobalRegistry.MustRegister(`gc.heap_obj`, NewGaugeFunc())
	stackInuse := GlobalRegistry.MustRegister(`gc.stack_inuse`, NewGaugeFunc("bytes"))
	mspanInuse := GlobalRegistry.MustRegister(`gc.mspan_inuse`, NewGaugeFunc("bytes"))
	mcacheInuse := GlobalRegistry.MustRegister(`gc.mcache_inuse`, NewGaugeFunc("bytes"))
	go func() {
		stats := &runtime.MemStats{}
		for {
			runtime.ReadMemStats(stats)
			gcCount.Update(stats.NumGC)
			gcPause.Update(WrapUint64Counter(stats.PauseTotalNs))
			gcCPUPercentage.Update(stats.GCCPUFraction * 100)
			mallocCount.Update(WrapUint64Counter(stats.Mallocs))
			freeCount.Update(WrapUint64Counter(stats.Frees))
			heapAlloc.Update(stats.HeapAlloc)
			heapIdle.Update(stats.HeapIdle)
			heapInuse.Update(stats.HeapInuse)
			heapObj.Update(stats.HeapObjects)
			stackInuse.Update(stats.StackInuse)
			mspanInuse.Update(stats.MSpanInuse)
			mcacheInuse.Update(stats.MCacheInuse)
			time.Sleep(interval)
		}
	}()
}
