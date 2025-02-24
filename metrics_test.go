package mon

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"

	"testing"
	"time"
)

func TestEWMA(t *testing.T) {
	ewma := NewEWMA(time.Minute, "")
	ewma.Update(12)
	assert.InDelta(t, 12, ewma.Value(), 0.1)
	m, err := json.Marshal(ewma)
	assert.Nil(t, err)
	assert.Equal(t, `{"type":"G","value":12}`, string(m), "EWMA serialization")
	ewma.Update(-11)
	assert.Less(t, ewma.Value(), float64(12), "EWMA update")

}

func TestEWMARate(t *testing.T) {
	ewma := NewEWMARate(time.Minute)
	ewma.Update(1)
	ewma.Update(1)
	ewma.Update(1)
	assert.Greater(t, ewma.Value(), float64(0))
	m, err := json.Marshal(ewma)
	assert.Nil(t, err)
	assert.Contains(t, string(m), `{"type":"G","value":`, "serialization")
}

func TestCounter(t *testing.T) {
	ctr := NewCounter()
	ctr.Update(123)
	assert.EqualValues(t, 123, ctr.Value(), "counter init")
	m, err := json.Marshal(ctr)
	assert.Nil(t, err)
	assert.Equal(t, string(m), `{"type":"C","value":123}`, "counter serialization")
	ctr.Update(-11)
	assert.EqualValues(t, 112, ctr.Value(), "counter update")
	// trigger overflow
	ctr.Update(float64(5000000000000000000))
	ctr.Update(float64(5000000000000000000))
	assert.EqualValues(t, 0, ctr.Value(), "wrapover")
}

func TestGauge(t *testing.T) {
	gauge := NewGauge("testUnit")
	gauge.Update(323)
	assert.Equal(t, 323.0, gauge.Value())
	gauge.Update(123)
	assert.Equal(t, 123.0, gauge.Value())
	assert.Equal(t, "testUnit", gauge.Unit())
	gaugeUnitless := NewGauge()
	gaugeUnitless.Update(32)
	assert.Equal(t, 32.0, gaugeUnitless.Value())
	gaugeUnitless.Update(432)
	assert.Equal(t, 432.0, gaugeUnitless.Value())
	assert.Equal(t, "", gaugeUnitless.Unit())
}
