package mon

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetrics(t *testing.T) {
	t.Run("marshal gauge", func(t *testing.T) {
		gauge := NewRawGauge()
		err1 := gauge.Update(0.3)
		assert.Nil(t, err1)
		err2 := gauge.Update(0.1)
		assert.Nil(t, err2)
		m, errJson := json.Marshal(gauge)
		assert.Nil(t, errJson)
		assert.Equal(t, `{"type":"G","value":0.1}`, string(m))
	})
	t.Run("marshal integer gauge", func(t *testing.T) {
		gaugeint := NewRawGaugeInt("cats")
		err1 := gaugeint.Update(321)
		assert.Nil(t, err1)
		err2 := gaugeint.Update(123)
		assert.Nil(t, err2)
		m, errJson := json.Marshal(gaugeint)
		assert.Nil(t, errJson)
		assert.Equal(t, `{"type":"g","unit":"cats","value":123}`, string(m))
	})
	t.Run("marshal counter", func(t *testing.T) {
		counter := NewRawCounter()
		err1 := counter.Update(4)
		assert.Nil(t, err1)
		err2 := counter.Update(1234)
		assert.Nil(t, err2)
		m, errJson := json.Marshal(counter)
		assert.Nil(t, errJson)
		assert.Equal(t, `{"type":"c","value":1234}`, string(m))
	})
}
