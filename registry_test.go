package mon

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

func TestRegisterGcStats(t *testing.T) {

	RegisterGcStats()
	m, err := json.Marshal(GlobalRegistry)
	assert.Nil(t, err)

	assert.Contains(t, string(m), `"instance"`)
	assert.Contains(t, string(m), `"go_gc_count"`)
}

func TestRegistry(t *testing.T) {
	GlobalRegistry.SetFQDN("asd.example.com")
	GlobalRegistry.SetInterval(10.0)
	GlobalRegistry.SetInstance("testInstance")
	assert.Equal(t, 10.0, GlobalRegistry.Interval)
	assert.Equal(t, "testInstance", GlobalRegistry.Instance)
	assert.Equal(t, "asd.example.com", GlobalRegistry.FQDN)
	_, err := GlobalRegistry.GetMetric("asd")
	assert.Error(t, err)
}

func BenchmarkRegistry_GetRegistry(b *testing.B) {
	for n := 0; n < b.N; n++ {
		GlobalRegistry.GetRegistry()
	}
}

func BenchmarkMemstat(b *testing.B) {
	stats := &runtime.MemStats{}
	for n := 0; n < b.N; n++ {
		runtime.ReadMemStats(stats)
	}
}

func BenchmarkGobtag(b *testing.B) {
	v := GobTag{T: map[string]string{
		"a": "b",
		"c": "d",
	}}
	for n := 0; n < b.N; n++ {
		gobTag(v)
	}
}
