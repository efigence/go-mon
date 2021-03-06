package mon

import (
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"testing"
	"time"
)

func TestEWMA(t *testing.T) {
	ewma := NewEWMA(time.Minute, "")
	errUpd1 := ewma.Update(12)
	Convey("EWMA Init", t, func() {
		So(errUpd1, ShouldEqual, nil)
		So(ewma.Value(), ShouldBeBetweenOrEqual, 11.9, 12)
	})
	m, err := json.Marshal(ewma)
	Convey("EWMA serialization", t, func() {
		So(err, ShouldBeNil)
		So(string(m), ShouldEqual, `{"type":"G","value":12}`)
	})
	errUpd2 := ewma.Update(-11)
	Convey("EWMA Update", t, func() {
		So(errUpd2, ShouldEqual, nil)
		So(ewma.Value(), ShouldBeLessThan, 12)
	})

}

func TestEWMARate(t *testing.T) {
	ewma := NewEWMARate(time.Minute)
	errUpd1 := ewma.Update(1)
	ewma.Update(1)
	ewma.Update(1)
	Convey("EWMA Init", t, func() {
		So(errUpd1, ShouldEqual, nil)
		So(ewma.Value(), ShouldBeGreaterThan, 0)
	})
	m, err := json.Marshal(ewma)
	Convey("EWMA serialization", t, func() {
		So(err, ShouldBeNil)
		So(string(m), ShouldContainSubstring, `{"type":"G","value":`)
	})
}

func TestCounter(t *testing.T) {
	ctr := NewCounter()
	ctr.Update(123)
	Convey("Counter init", t, func() {
		So(ctr.ValueRaw(), ShouldEqual, 123)
	})
	m, err := json.Marshal(ctr)
	Convey("Counter serialization", t, func() {
		So(err, ShouldBeNil)
		So(string(m), ShouldEqual, `{"type":"c","value":123}`)
	})
	errUpd2 := ctr.Update(-11)
	Convey("counter update", t, func() {
		So(errUpd2, ShouldEqual, nil)
		So(ctr.Value(), ShouldEqual, 112)
	})
	// trigger overflow
	errUpd3 := ctr.Update(5000000000000000000)
	errUpd4 := ctr.Update(5000000000000000000)
	Convey("counter wrapover", t, func() {
		So(errUpd2, ShouldEqual, nil)
		So(errUpd3, ShouldEqual, nil)
		So(errUpd4, ShouldEqual, nil)
		So(ctr.Value(), ShouldEqual, 0)
	})
}

func TestGauge(t *testing.T) {
	gauge := NewGauge("testUnit")
	gauge.Update(323)
	assert.Equal(t,323.0,gauge.Value(),)
	gauge.Update(123)
	assert.Equal(t,123.0,gauge.Value(),)
	assert.Equal(t,"testUnit",gauge.Unit())
	gaugeUnitless := NewGauge()
	gaugeUnitless.Update(32)
	assert.Equal(t,32.0,gaugeUnitless.Value(),)
	gaugeUnitless.Update(432)
	assert.Equal(t,432.0,gaugeUnitless.Value(),)
	assert.Equal(t,"",gaugeUnitless.Unit())

}
