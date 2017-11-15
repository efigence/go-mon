package mon

import (
)

// NewRawCounter return raw (no backend) counter.
func NewRawCounter(i int64) Metric {
	return &MetricInt {
		metricType: MetricTypeCounter,
		value: i,
	}

}

// NewRawCounter return raw (no backend) counter.
func NewRawCounterFloat(f float64) Metric {
	return &MetricFloat {
		metricType: MetricTypeCounterFloat,
		value: f,
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
