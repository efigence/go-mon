package mon

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
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
	Update(interface{}) error
	Unit() string
	Value() float64
	ValueRaw() interface{}
	//	json.Marshaler
}

// Return int64 or conversion error
func Int64OrError(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case uint32: // uint64 is skipped because it doesn't fit in int64
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("Got type %T, expected int64 or any smaller one that fits", value)
	}
}

// Return float64 or conversion error
func Float64OrError(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	default:
		return math.NaN(), fmt.Errorf("Got type %T, expected float64 or any smaller one that fits", value)
	}
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
	metricType string  `json:"type"`
	unit       string  `json:"unit,omitempty"`
	value      float64 `json:"value"`
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

func (f *MetricFloat) Update(value interface{}) (err error) {
	v, err := Float64OrError(value)
	f.Lock()
	if err == nil {
		f.value = v
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

// raw int metric with no backend
type MetricInt struct {
	metricType string
	unit       string
	value      int64
	sync.Mutex
}

func (f *MetricInt) Type() string {
	return f.metricType
}
func (f *MetricInt) Unit() string {
	return f.unit
}
func (f *MetricInt) Value() float64 {
	return float64(atomic.LoadInt64(&f.value))
}
func (f *MetricInt) ValueRaw() interface{} {
	return atomic.LoadInt64(&f.value)
}
func (f *MetricInt) Update(value interface{}) (err error) {
	v, err := Int64OrError(value)
	if err == nil {
		// ignored on purpose; if 2 writes happen at same time there is no "right" answer whether to repeat or not
		atomic.CompareAndSwapInt64(&f.value, f.value, v)
	}
	return err
}
func (f *MetricInt) MarshalJSON() ([]byte, error) {
	f.Lock()
	defer f.Unlock()
	return json.Marshal(
		JSONOut{
			Type:  f.metricType,
			Value: f.value,
			Unit:  f.unit,
		})
}

// Int metric with backend
//
// By default backend is updated with mutex lock, all other locking have to be
// handled by the backend itself

type MetricIntBackend struct {
	metricType string
	unit       string
	backend    StatBackendInt
	sync.Mutex
}

func (f *MetricIntBackend) Type() string {
	return f.metricType
}
func (f *MetricIntBackend) Unit() string {
	return f.unit
}
func (f *MetricIntBackend) Value() float64 {
	return float64(f.backend.Value())
}
func (f *MetricIntBackend) ValueRaw() interface{} {
	return f.backend.Value()
}
func (f *MetricIntBackend) MarshalJSON() ([]byte, error) {
	f.Lock()
	defer f.Unlock()
	return json.Marshal(
		JSONOut{
			Type:  f.metricType,
			Unit:  f.unit,
			Value: f.backend.Value(),
		})
}
func (f *MetricIntBackend) Update(value interface{}) (err error) {
	v, err := Int64OrError(value)
	f.Lock()
	if err == nil {
		f.backend.Update(v)
	}
	f.Unlock()
	return err
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
func (f *MetricFloatBackend) ValueRaw() interface{} {
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
func (f *MetricFloatBackend) Update(value interface{}) (err error) {
	v, err := Float64OrError(value)
	f.Lock()
	if err == nil {
		f.backend.Update(v)
	}
	f.Unlock()
	return err
}
