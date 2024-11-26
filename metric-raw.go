package mon

// NewRawCounter return raw (no backend) counter.
func NewRawCounter(unit ...string) Metric {
	u := ""
	if len(unit) > 0 {
		u = unit[0]
	}
	return &MetricInt{
		metricType: MetricTypeCounter,
		unit:       u,
	}

}

// NewRawCounter return raw (no backend) counter.
func NewRawCounterFloat(unit ...string) Metric {
	u := ""
	if len(unit) > 0 {
		u = unit[0]
	}
	return &MetricFloat{
		metricType: MetricTypeCounterFloat,
		unit:       u,
	}

}

// NewRawCounter return raw (no backend) integer gauge.
func NewRawGaugeInt(unit ...string) Metric {
	u := ""
	if len(unit) > 0 {
		u = unit[0]
	}
	return &MetricInt{
		metricType: MetricTypeGaugeInt,
		unit:       u,
	}

}

// NewRawCounter return raw (no backend) float gauge.
func NewRawGauge(unit ...string) Metric {
	u := ""
	if len(unit) > 0 {
		u = unit[0]
	}
	return &MetricFloat{
		metricType: MetricTypeGauge,
		unit:       u,
	}

}
