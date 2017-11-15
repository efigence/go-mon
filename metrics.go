package mon
import (
	"github.com/efigence/go-libs/ewma"
	"time"
	"sync/atomic"
)

type ewmaBackend struct {
	ewma *ewma.Ewma
}

func (e *ewmaBackend)Update(u float64) {
	e.ewma.UpdateNow(u)
}
func (e *ewmaBackend)Value() float64 {
	return e.ewma.Current
}

// New exponentally weighted moving average metric
func NewEWMA(halflife time.Duration) Metric {
	return &MetricFloatBackend{
		metricType: MetricTypeGauge,
		backend:  &ewmaBackend{
			ewma: ewma.NewEwma(halflife),
		},
	}
}


type counterBackend struct {
	counter int64
	canBeNegative bool
}
func (c *counterBackend)Update(u int64) {
	atomic.AddInt64(&c.counter,u)
	// FIXME probably should add remainder from overflow instead of zeroing it
	if !c.canBeNegative && c.counter < 0 {
		ctr := c.counter
		for ctr < 0 {
			atomic.CompareAndSwapInt64(&c.counter, c.counter, 0)
			ctr = c.counter
		}
	}
}
func (c *counterBackend)Value() int64 {
	ctr := c.counter
	if !c.canBeNegative && ctr < 0 {
		return 0
	}
	return ctr
}
// New counter. Updating it will INCREMENT internal counter. If you want to just set a value directly, use NewRawCounter() instead. Overflows to zero
func NewCounter() Metric {
	return &MetricIntBackend {
		metricType: MetricTypeCounter,
		backend: &counterBackend{},
	}
}
