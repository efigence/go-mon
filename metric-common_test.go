package mon

import (
	"testing"
		. "github.com/smartystreets/goconvey/convey"
	"reflect"
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
	Convey("to float64", t, func() {

		for _, t := range testsFloat {
			_ = t
			f, err := Float64OrError(t)
			Convey("from " + reflect.TypeOf(t).Name(), func() {
				So(err, ShouldBeNil)
				So(f, ShouldEqual, float64(10))
			})
		}
	})
	_, err := Float64OrError("sdasd")
	Convey("from string to float64", t, func() {
		So(err, ShouldNotBeNil)
	})
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
	Convey("to int64", t, func() {

		for _, t := range testsInt {
			_ = t
			f, err := Int64OrError(t)
			Convey("from " + reflect.TypeOf(t).Name(), func() {
				So(err, ShouldBeNil)
				So(f, ShouldEqual, float64(10))
			})
		}
	})
	_, err = Int64OrError("sdasd")
	Convey("from string to int64", t, func() {
		So(err, ShouldNotBeNil)
	})
	_, err = Int64OrError(uint64(10))
	Convey("from uint64 to int64", t, func() {
		So(err, ShouldNotBeNil)
	})
	_, err = Int64OrError(float32(10))
	Convey("from string to int64", t, func() {
		So(err, ShouldNotBeNil)
	})


}