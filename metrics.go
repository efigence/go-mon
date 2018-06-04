package mon

import (
	"github.com/efigence/go-libs/ewma"
	"sync/atomic"
	"time"
)

type ewmaBackend struct {
	ewma *ewma.Ewma
}

func (e *ewmaBackend) Update(u float64) {
	e.ewma.UpdateNow(u)
}
func (e *ewmaBackend) Value() float64 {
	return e.ewma.Current
}

// New exponentally weighted moving average metric
// halflife is half-life of stat decay
//
func NewEWMA(halflife time.Duration, unit ...string) Metric {
	metric := &MetricFloatBackend{
		metricType: MetricTypeGauge,
		backend: &ewmaBackend{
			ewma: ewma.NewEwma(halflife),
		},
	}
	if len(unit) > 0 {
		metric.unit = unit[0]
	}
	return metric
}

type ewmaRateBackend struct {
	ewma *ewma.EwmaRate
}
// Update rate counter. Value is ignored, each Update() call is one "request" for rate calculation
func (e *ewmaRateBackend) Update(u float64) {
	e.ewma.UpdateNow()
}
func (e *ewmaRateBackend) Value() float64 {
	return e.ewma.CurrentNow()
}

// New exponentally weighted moving average event rate counter
// call Update(value is ignored) every time an event happens to get rate of the event
//
func NewEWMARate(halflife time.Duration, unit ...string) Metric {
	metric := &MetricFloatBackend{
		metricType: MetricTypeGauge,
		backend: &ewmaRateBackend{
			ewma: ewma.NewEwmaRate(halflife),
		},
	}
	if len(unit) > 0 {
		metric.unit = unit[0]
	}
	return metric
}

type counterBackend struct {
	counter       int64
	canBeNegative bool
}

func (c *counterBackend) Update(u int64) {
	atomic.AddInt64(&c.counter, u)
	// FIXME probably should add remainder from overflow instead of zeroing it
	if !c.canBeNegative && c.counter < 0 {
		ctr := c.counter
		for ctr < 0 {
			atomic.CompareAndSwapInt64(&c.counter, c.counter, 0)
			ctr = c.counter
		}
	}
}
func (c *counterBackend) Value() int64 {
	ctr := atomic.LoadInt64(&c.counter)
	if !c.canBeNegative && ctr < 0 {
		return 0
	}
	return ctr
}

// New counter. Updating it will INCREMENT internal counter. If you want to just set a value directly, use NewRawCounter() instead. Overflows to zero
func NewCounter() Metric {
	return &MetricIntBackend{
		metricType: MetricTypeCounter,
		backend:    &counterBackend{},
	}
}
