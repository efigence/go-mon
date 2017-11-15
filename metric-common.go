package mon


import (
	"encoding/json"
	"fmt"
	"math"
)

const (
	MetricTypeGauge = `G` // float64 gauge
	MetricTypeGaugeInt = `g` // int64 gauge
	MetricTypeCounter = `c` // int64 counter
	MetricTypeCounterFloat = `C` // float64 counter
)

type Metric interface {
	Type() string
	Update(interface{}) error
	Value() float64
	ValueRaw() interface{}
//	json.Marshaler
}

func Int64OrError(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case uint32: // uint64 is skipped because it doesn't fit in int64
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	default:
		return 0,fmt.Errorf("Got type %T, expected float64", value)
	}
}
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
	default:
		return math.NaN(),fmt.Errorf("Got type %T, expected float64", value)
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
	Type string `json:"type"`
	Value interface{} `json:"value"`
}

// raw float metric with no backend
type MetricFloat struct {
	metricType string `json:"type"`
	value float64 `json:"value"`
}

func (f *MetricFloat) Type() string {
	return f.metricType
}
func (f *MetricFloat) Value() float64 {
	return f.value
}
func (f *MetricFloat) ValueRaw() interface{} {
	return f.value
}

func (f *MetricFloat) Update(value interface{}) (err error) {
	v, err := Float64OrError(value)
	if err != nil { f.value = v }
	return err
}


func (f *MetricFloat) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		JSONOut {
			Type: f.metricType,
			Value: f.value,
		})
}

// raw int metric with no backend
type MetricInt struct {
	metricType string `json:"type"`
	value int64 `json:"value"`

}
func (f *MetricInt) Type() string {
	return f.metricType
}
func (f *MetricInt) Value() float64 {
	return float64(f.value)
}
func (f *MetricInt) ValueRaw() interface{} {
	return f.value
}
func (f *MetricInt) Update(value interface{}) (err error) {
	v, err := Int64OrError(value)
	if err != nil { f.value = v }
	return err
}
func (f *MetricInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		JSONOut {
			Type: f.metricType,
			Value: f.value,
		})
}
// Int metric with backend
type MetricIntBackend struct {
	metricType string `json:"type"`
	backend StatBackendInt
}
func (f *MetricIntBackend) Type() string {
	return f.metricType
}
func (f *MetricIntBackend) Value() float64 {
	return float64(f.backend.Value())
}
func (f *MetricIntBackend) ValueRaw() interface{} {
	return f.backend.Value()
}
func (f *MetricIntBackend) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		JSONOut {
			Type: f.metricType,
			Value: f.backend.Value(),
		})
}
func (f *MetricIntBackend) Update(value interface{}) (err error) {
	v, err := Int64OrError(value)
	if err == nil { f.backend.Update(v) }
	return err
}

// float metric with backend
type MetricFloatBackend struct {
	metricType string `json:"type"`
	backend StatBackendFloat
}
func (f *MetricFloatBackend) Type() string {
	return f.metricType
}
func (f *MetricFloatBackend) Value() float64 {
	return f.backend.Value()
}
func (f *MetricFloatBackend) ValueRaw() interface{} {
	return f.backend.Value()
}
func (f *MetricFloatBackend) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		JSONOut {
			Type: f.metricType,
			Value: f.backend.Value(),
		})
}
func (f *MetricFloatBackend) Update(value interface{}) (err error) {
	v, err := Float64OrError(value)
	if err == nil { f.backend.Update(v) }
	return err
}
