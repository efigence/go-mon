package mon

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"encoding/json"
)

func TestMetrics(t *testing.T) {
	gauge := NewRawGauge()
	err1 := gauge.Update(0.3)
	err2 := gauge.Update(0.1)
	m,errJson := json.Marshal(gauge)
	Convey("Marshal Gauge", t, func() {
		So(err1, ShouldEqual, nil)
		So(err2, ShouldEqual, nil)
		So(errJson, ShouldEqual, nil)
		So(string(m), ShouldEqual, `{"type":"G","value":0.1}`)
	})

	gaugeint := NewRawGaugeInt("cats")
	err1 = gaugeint.Update(321)
	err2 = gaugeint.Update(123)
	m,errJson = json.Marshal(gaugeint)
	Convey("Marshal Integer Gauge", t, func() {
		So(err1, ShouldEqual, nil)
		So(err2, ShouldEqual, nil)
		So(errJson, ShouldEqual, nil)
		So(string(m), ShouldEqual, `{"type":"g","unit":"cats","value":123}`)
	})

	counter := NewRawCounter()
	err1 = counter.Update(4)
	err2 = counter.Update(1234)
	m,errJson = json.Marshal(counter)
	Convey("Marshal Counter", t, func() {
		So(err1, ShouldEqual, nil)
		So(err2, ShouldEqual, nil)
		So(errJson, ShouldEqual, nil)
		So(string(m), ShouldEqual, `{"type":"c","value":1234}`)
	})
}
