package mon

import (
	"github.com/stretchr/testify/assert"
	"math"
	"reflect"
	"testing"
	"time"
)

func TestMetricsCommon(t *testing.T) {
	var testsFloat []interface{}
	testsFloat = append(testsFloat,
		int(10),
		int64(10),
		int32(10),
		int16(10),
		int8(10),
		uint(10),
		uint64(10),
		uint32(10),
		uint16(10),
		uint8(10),
		float32(10),
	)
	t.Run("to float", func(t *testing.T) {
		for _, num := range testsFloat {
			f, err := Float64OrError(num)
			typeName := reflect.TypeOf(num).Name()
			assert.Nil(t, err, typeName)
			assert.Equal(t, float64(10), f, typeName)
		}
	})

	_, err := Float64OrError("sdasd")
	assert.Error(t, err)

	var testsInt []interface{}
	testsInt = append(testsInt,
		int64(10),
		int32(10),
		int16(10),
		int8(10),
		uint32(10),
		uint16(10),
		uint8(10),
	)
	t.Run("to int64", func(t *testing.T) {
		for _, ti := range testsInt {
			_ = ti
			f, err := Int64OrError(ti)
			assert.Nil(t, err)
			assert.Equal(t, int64(10), f, reflect.TypeOf(ti).Name())
		}
	})

	_, err = Int64OrError("sdasd")
	assert.Error(t, err)
	_, err = Int64OrError(uint64(10))
	assert.Error(t, err)
	_, err = Int64OrError(float32(10))
	assert.Error(t, err)

}

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

func BenchmarkMetricInt_Update(b *testing.B) {
	r := NewRawGaugeInt()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r.Update(1)
	}
}
func BenchmarkMetricInt_Value(b *testing.B) {
	r := NewRawGaugeInt()
	r.Update(1)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = r.Value()
	}
}
func BenchmarkMetricFloat_Update(b *testing.B) {
	r := NewRawGauge()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r.Update(1)
	}
}
func BenchmarkMetricFloat_Value(b *testing.B) {
	r := NewRawGauge()
	r.Update(1)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = r.Value()
	}
}
