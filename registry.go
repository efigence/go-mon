package mon

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"runtime"
	"sync"
	"time"
)

var emptyGob = gobTag(GobTag{map[string]string{}})

type Registry struct {
	Metrics  map[string]map[string]Metric `json:"metrics"`
	Instance string                       `json:"instance"`
	Interval float64                      `json:"interval"`
	FQDN     string                       `json:"fqdn"`
	Ts       time.Time                    `json:"ts,omitempty"`
	sync.Mutex
}

func NewRegistry(fqdn string, instance string, interval float64) (*Registry, error) {
	return &Registry{
		FQDN:     fqdn,
		Instance: instance,
		Interval: interval,
		Metrics:  make(map[string]map[string]Metric),
	}, nil
}

type GobTag struct {
	T map[string]string
}

func mapToGobTag(v ...map[string]string) GobTag {
	d := map[string]string{}
	for _, m := range v {
		for k, v := range m {
			d[k] = v
		}
	}
	return GobTag{T: d}
}

func gobTag(g GobTag) (data []byte) {
	var b bytes.Buffer
	err := gob.NewEncoder(&b).Encode(&g)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}
func gobUntag(data []byte) (g GobTag) {
	b := bytes.NewReader(data)
	err := gob.NewDecoder(b).Decode(&g)
	if err != nil {
		panic(err)
	}
	return g
}

func (r *Registry) GetMetric(name string, tags ...map[string]string) (Metric, error) {
	gob := gobTag(mapToGobTag(tags...))

	if r, ok := r.Metrics[name]; ok {
		if r, ok := r[string(gob)]; ok {
			return r, nil
		} else {
			return nil, &ErrMetricNotFound{Metric: name}
		}
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
		Metrics:  make(map[string]map[string]Metric),
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
func (r *Registry) Register(name string, metric Metric, tags ...map[string]string) (Metric, error) {
	gob := gobTag(mapToGobTag(tags...))
	r.Lock()
	defer r.Unlock()
	if _, ok := r.Metrics[name]; !ok {
		r.Metrics[name] = make(map[string]Metric, 0)
	}
	if r, ok := r.Metrics[name][string(gob)]; ok {
		return r, &ErrMetricAlreadyRegistered{Metric: name}
	}
	r.Metrics[name][string(gob)] = metric
	return metric, nil
}

// RegisterOrGet() registers a given metric or resturns already existing one if it is of same type
// it will err out if type does not match but it does not compare rest of the parameters of the metric so do not use it if you are not 100% sure

func (r *Registry) RegisterOrGet(name string, metric Metric, tags ...map[string]string) (Metric, error) {
	gob := gobTag(mapToGobTag(tags...))
	r.Lock()
	defer r.Unlock()
	if _, ok := r.Metrics[name]; !ok {
		r.Metrics[name] = make(map[string]Metric, 0)
	}
	if m, ok := r.Metrics[name][string(gob)]; ok {
		if m.Type() == metric.Type() {
			return m, nil
		} else {
			return m, &ErrMetricAlreadyRegisteredWrongType{
				Metric:        name,
				OldMetricType: m.Type(),
				NewMetricType: metric.Type(),
			}
		}
	} else {
		r.Metrics[name][string(gob)] = metric
	}
	return metric, nil
}

// MustRegister() does same as Register() except it panic()s if metric already exists.
// It is mostly intended to be used for top of the package, package-scoped metrics like
//
//	var request_rate =  mon.GlobalRegistry.Register("backend.mysql.qps",mon.NewEWMARate(time.Duration(time.Minute)))
func (r *Registry) MustRegister(name string, metric Metric, tags ...map[string]string) Metric {
	m, err := r.Register(name, metric, tags...)
	if err != nil {
		panic(fmt.Sprintf("Failed to register metric %s: %s", name, err))
	}
	return m
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
		return NewGauge(unit...)
	}
	if average {
		NewGaugeFunc = func(unit ...string) Metric {
			return NewEWMA(EWMAHalfLife, unit...)
		}
	}
	gcCount := GlobalRegistry.MustRegister(`gc.count`, NewRawCounter())
	gcPause := GlobalRegistry.MustRegister(`gc.pause`, NewRawCounter("ns"))
	gcCPUPercentage := GlobalRegistry.MustRegister(`gc.cpu`, NewGauge("percent"))
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
			gcCount.Update(float64(stats.NumGC))
			gcPause.Update(float64(stats.PauseTotalNs))
			gcCPUPercentage.Update(stats.GCCPUFraction * 100)
			mallocCount.Update(float64(stats.Mallocs))
			freeCount.Update(float64(stats.Frees))
			heapAlloc.Update(float64(stats.HeapAlloc))
			heapIdle.Update(float64(stats.HeapIdle))
			heapInuse.Update(float64(stats.HeapInuse))
			heapObj.Update(float64(stats.HeapObjects))
			stackInuse.Update(float64(stats.StackInuse))
			mspanInuse.Update(float64(stats.MSpanInuse))
			mcacheInuse.Update(float64(stats.MCacheInuse))
			time.Sleep(interval)
		}
	}()
}
