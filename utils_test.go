package mon

import (
	"github.com/stretchr/testify/assert"
	"math"
	"os"
	"testing"
)

func TestGetFQDN(t *testing.T) {
	hostname, _ := os.Hostname()

	// that would probably fail on misconfigured machine... not my problem, fix your shit
	assert.Contains(t, getFQDN(), ".", "should contain fqdn with dots")
	assert.Contains(t, getFQDN(), hostname, "should contain its own hostname")
}

func TestWrapUint64Counter(t *testing.T) {
	under := uint64(math.MaxInt64 - 1)
	edge := uint64(math.MaxInt64)
	over0 := uint64(math.MaxInt64 + 1)
	over10 := uint64(math.MaxInt64 + 1 + 10)
	top := uint64(math.MaxUint64)
	assert.Equal(t, int64(math.MaxInt64-1), WrapUint64Counter(under), "1 away from overflow")
	assert.Equal(t, int64(math.MaxInt64), WrapUint64Counter(edge), "at overflow")
	assert.Equal(t, int64(0), WrapUint64Counter(over0), "1 overflow")
	assert.Equal(t, int64(10), WrapUint64Counter(over10), "10 overflow")
	assert.Equal(t, int64(math.MaxInt64), WrapUint64Counter(top), "uint64 overflow")
}
