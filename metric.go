package mon

import (
	"encoding/json"
)


const (
	MetricTypeGauge = `G` // float64 gauge
	MetricTypeGaugeInt = `g` // int64 gauge
	MetricTypeCounter = `c` // int64 counter
)



type Metric interface {
	Type() string
//	json.Marshaler
}

type MetricsBackendInt interface {
	Update(int64)
	Value(int64)
}

type MetricsBackendFloat interface {
	Update(float64)
	Value(float64)
}

type JSONOut struct {
	Type string `json:"type"`
	Value interface{} `json:"value"`
}

type MetricFloat struct {
	metricType string `json:"type"`
	value float64 `json:"value"`
	backend MetricsBackendFloat
}

type MetricInt struct {
	metricType string `json:"type"`
	value int64 `json:"value"`
	backend MetricsBackendInt

}

type MetricFloatBackend struct {
	metricType string `json:"type"`
	backend MetricsBackendFloat
}

type MetricIntBackend struct {
	metricType string `json:"type"`
	backend MetricsBackendInt
}

func (f *MetricFloat) Type() string {
	return f.metricType
}

func (f *MetricFloat) Value() string {
	return f.metricType
}

func (f *MetricFloat) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		JSONOut {
			Type: f.metricType,
			Value: f.value,
		})
}
func (f *MetricInt) Type() string {
	return f.metricType
}

func (f *MetricInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		JSONOut {
			Type: f.metricType,
			Value: f.value,
		})
}
// NewRawCounter return raw (no backend) counter.
func NewRawCounter(i int64) Metric {
	return &MetricInt {
		metricType: MetricTypeCounter,
		value: i,
	}

}
// NewRawCounter return raw (no backend) integer gauge.
func NewRawGaugeInt(i int64) Metric {
	return &MetricInt {
		metricType: MetricTypeGaugeInt,
		value: i,
	}

}
// NewRawCounter return raw (no backend) float gauge.
func NewRawGauge(f float64) Metric {
	return &MetricFloat {
		metricType: MetricTypeGauge,
		value: f,
	}

}
