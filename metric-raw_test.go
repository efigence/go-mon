package mon

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"encoding/json"
)

func TestMetrics(t *testing.T) {
	gauge := NewRawGauge(0.1)
	m,err := json.Marshal(gauge)
	Convey("Marshal Gauge", t, func() {
		So(err, ShouldEqual, nil)
		So(string(m), ShouldEqual, `{"type":"G","value":0.1}`)
	})

	gaugeint := NewRawGaugeInt(123)
	m,err = json.Marshal(gaugeint)
	Convey("Marshal Integer Gauge", t, func() {
		So(err, ShouldEqual, nil)
		So(string(m), ShouldEqual, `{"type":"g","value":123}`)
	})

	counter := NewRawCounter(1234)
	m,err = json.Marshal(counter)
	Convey("Marshal Counter", t, func() {
		So(err, ShouldEqual, nil)
		So(string(m), ShouldEqual, `{"type":"c","value":1234}`)
	})
}
