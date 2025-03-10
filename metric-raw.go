package mon

import (
	"encoding/json"
	"math"
	"sync"
)

type MetricRawCounter struct {
	value float64
	unit  string
	tags  map[string]string
	lock  sync.RWMutex
}

func (m *MetricRawCounter) Type() string {
	return MetricTypeCounterFloat
}
func (m *MetricRawCounter) Update(v float64) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.value = v
}
func (m *MetricRawCounter) Unit() string {
	return m.unit
}
func (m *MetricRawCounter) Value() float64 {
	// precison around 1e25 is above 0.1 so we reset counter back to 0
	// technically integer gapless range is up to 2^53 (~ 9E15) but I want some leeway here
	if m.value > 1e15 || m.value < (-1e15) {
		m.lock.Lock()
		m.value = 0
		m.lock.Unlock()
	}
	m.lock.RLock()
	defer m.lock.RUnlock()
	return float64(m.value)
}

func (f *MetricRawCounter) MarshalJSON() ([]byte, error) {
	// Go bug #3480 #25721
	// returning number is only option, or else Go (or other strict deserializers) will crap out on ingestion
	if math.IsNaN(f.value) {
		return json.Marshal(
			JSONOut{
				Type:    MetricTypeCounterFloat,
				Invalid: true,
				Unit:    f.unit,
			})
	}
	return json.Marshal(
		JSONOut{
			Type:  MetricTypeCounterFloat,
			Value: f.value,
			Unit:  f.unit,
		})
}

func NewRawCounter(unit ...string) Metric {
	m := MetricRawCounter{}
	if len(unit) > 0 {
		m.unit = unit[0]
	}
	return &m
}
