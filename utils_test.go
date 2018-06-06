package mon

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"math"
)
func TestGetFQDN(t *testing.T) {
	hostname,_ := os.Hostname()
	Convey("ShouldReturnAllMessages", t, func() {
		// TODO that would probably fail on misconfigured machine... not my problem, fix your shit
		So(getFQDN(), ShouldContainSubstring, ".")
		So(getFQDN(), ShouldContainSubstring, hostname)
	})
}

func TestWrapUint64Counter(t *testing.T) {
	under := uint64(math.MaxInt64 - 1)
	edge := uint64(math.MaxInt64)
	over0 := uint64(math.MaxInt64 + 1)
	over10 := uint64(math.MaxInt64 + 1 + 10)
	top := uint64(math.MaxUint64)
	Convey("1:1 below or at overflow", t, func() {
		So(WrapUint64Counter(under),ShouldEqual,math.MaxInt64 - 1)
		So(WrapUint64Counter(edge),ShouldEqual,math.MaxInt64)
		So(WrapUint64Counter(over0),ShouldEqual,0)
	})
	Convey("overflow", t, func() {
		So(WrapUint64Counter(over10),ShouldEqual,10)
		So(WrapUint64Counter(top),ShouldEqual, math.MaxInt64)

	})
}