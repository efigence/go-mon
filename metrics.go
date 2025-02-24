package mon

import (
	"github.com/efigence/go-libs/ewma"
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
