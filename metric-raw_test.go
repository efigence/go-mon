package mon

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetrics(t *testing.T) {
	t.Run("marshal counter", func(t *testing.T) {
		counter := NewCounter()
		counter.Update(4)
		counter.Update(1234)
		m, errJson := json.Marshal(counter)
		assert.Nil(t, errJson)
		assert.Equal(t, `{"type":"C","value":1238}`, string(m))
	})
	t.Run("marshal raw counter", func(t *testing.T) {
		counter := NewRawCounter()
		counter.Update(4)
		counter.Update(1234)
		m, errJson := json.Marshal(counter)
		assert.Nil(t, errJson)
		assert.Equal(t, `{"type":"C","value":1234}`, string(m))
	})
}
