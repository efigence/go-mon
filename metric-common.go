package mon

import (
	"encoding/json"
	"math"
	"sync"
)

const (
	MetricTypeGauge        = `G` // float64 gauge
	MetricTypeGaugeInt     = `g` // int64 gauge
	MetricTypeCounter      = `c` // int64 counter
	MetricTypeCounterFloat = `C` // float64 counter
)

// Single metric handler interface
type Metric interface {
	Type() string
	Update(float64)
	Unit() string
	Value() float64
	//	json.Marshaler
}
type MetricGauge struct {
	value float64
	unit  string
	lock  sync.RWMutex
}

func (m *MetricGauge) Type() string {
	return MetricTypeGauge
}
func (m *MetricGauge) Update(v float64) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.value = v
}
func (m *MetricGauge) Unit() string {
	return m.unit
}
func (m *MetricGauge) Value() float64 {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return float64(m.value)
}

func NewGauge(unit ...string) Metric {
	m := MetricGauge{}
	if len(unit) > 0 {
		m.unit = unit[0]
	}
	return &m
}
func (f *MetricGauge) MarshalJSON() ([]byte, error) {
	// Go bug #3480 #25721
	// returning number is only option, or else Go (or other strict deserializers) will crap out on ingestion
	if math.IsNaN(f.value) {
		return json.Marshal(
			JSONOut{
				Type:    MetricTypeGauge,
				Invalid: true,
				Unit:    f.unit,
			})
	}
	return json.Marshal(
		JSONOut{
			Type:  MetricTypeGauge,
			Value: f.value,
			Unit:  f.unit,
		})
}

type MetricCounter struct {
	value float64
	unit  string
	lock  sync.RWMutex
}

func (m *MetricCounter) Type() string {
	return MetricTypeCounterFloat
}
func (m *MetricCounter) Update(v float64) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.value += v
}
func (m *MetricCounter) Unit() string {
	return m.unit
}
func (m *MetricCounter) Value() float64 {
	// precison around 1e25 is above 0.1 so we reset counter back to 0
	if m.value > 1e15 || m.value < (-1e15) {
		m.lock.Lock()
		m.value = 0
		m.lock.Unlock()
	}
	m.lock.RLock()
	defer m.lock.RUnlock()
	return float64(m.value)
}

func NewCounter(unit ...string) Metric {
	m := MetricCounter{}
	if len(unit) > 0 {
		m.unit = unit[0]
	}
	return &m
}
func (f *MetricCounter) MarshalJSON() ([]byte, error) {
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

// backend interface handling single integer stat
type StatBackendInt interface {
	Update(int64)
	Value() int64
}

// backend interface handling single integer stat
type StatBackendFloat interface {
	Update(float64)
	Value() float64
}

// API-compatible JSON output structure
type JSONOut struct {
	Type    string      `json:"type"`
	Unit    string      `json:"unit,omitempty"`
	Invalid bool        `json:"invalid,omitempty"`
	Value   interface{} `json:"value"`
}

// raw float metric with no backend
type MetricFloat struct {
	metricType string
	unit       string
	value      float64
	sync.Mutex
}

func (f *MetricFloat) Type() string {
	return f.metricType
}
func (f *MetricFloat) Unit() string {
	return f.unit
}
func (f *MetricFloat) Value() float64 {
	return f.value
}
func (f *MetricFloat) ValueRaw() interface{} {
	return f.value
}

func (f *MetricFloat) Update(value float64) (err error) {
	f.Lock()
	if err == nil {
		f.value = value
	}
	f.Unlock()
	return err
}

func (f *MetricFloat) MarshalJSON() ([]byte, error) {
	// Go bug #3480 #25721
	// returning number is only option, or else Go (or other strict deserializers) will crap out on ingestion
	if math.IsNaN(f.value) {
		return json.Marshal(
			JSONOut{
				Type:    f.metricType,
				Invalid: true,
				Unit:    f.unit,
			})
	}
	return json.Marshal(
		JSONOut{
			Type:  f.metricType,
			Value: f.value,
			Unit:  f.unit,
		})
}

// Float metric with backend.
//
// By default backend is updated with mutex lock, all other locking have to be
// handled by the backend itself
type MetricFloatBackend struct {
	metricType string
	unit       string
	backend    StatBackendFloat
	sync.Mutex
}

func (f *MetricFloatBackend) Type() string {
	return f.metricType
}
func (f *MetricFloatBackend) Unit() string {
	return f.unit
}
func (f *MetricFloatBackend) Value() float64 {
	return f.backend.Value()
}

func (f *MetricFloatBackend) MarshalJSON() ([]byte, error) {
	f.Lock()
	defer f.Unlock()
	v := f.backend.Value()
	// Go bug #3480 #25721
	// returning number is only option, or else Go (or other strict deserializers) will crap out on ingestion
	if math.IsNaN(v) {
		return json.Marshal(JSONOut{
			Type:    f.metricType,
			Unit:    f.unit,
			Invalid: true,
		})
	}
	return json.Marshal(
		JSONOut{
			Type:  f.metricType,
			Unit:  f.unit,
			Value: v,
		})
}
func (f *MetricFloatBackend) Update(value float64) {
	f.Lock()
	f.backend.Update(value)
	defer f.Unlock()
}
