package mon

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
	"time"
)

func TestMetricsNaN(t *testing.T) {
	g := MetricFloat{
		metricType: "g",
		value:      math.NaN(),
	}

	out, err := g.MarshalJSON()
	assert.Nil(t, err)
	assert.Contains(t, string(out), `"invalid":true`)

}

func BenchmarkEwmaBackend_Update(b *testing.B) {
	r := NewEWMARate(time.Minute)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r.Update(1)
	}
}

func BenchmarkEwmaBackend_Value(b *testing.B) {
	r := NewEWMARate(time.Minute)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = r.Value()
	}
}

func BenchmarkMetricGauge_Update(b *testing.B) {
	r := NewGauge()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r.Update(1)
	}
}
func BenchmarkMetricGauge_Value_Value(b *testing.B) {
	r := NewGauge()
	r.Update(1)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = r.Value()
	}
}
func BenchmarkMetricCounter_Update(b *testing.B) {
	r := NewCounter()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r.Update(1)
	}
}
func BenchmarkMetricCounter_Value(b *testing.B) {
	r := NewCounter()
	r.Update(1)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = r.Value()
	}
}
