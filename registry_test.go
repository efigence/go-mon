package mon

import (
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDummy(t *testing.T) {

	RegisterGcStats()
	Convey("Marshal registry", t, func() {
		m, err := json.Marshal(GlobalRegistry)
		So(err, ShouldBeNil)
		So(string(m), ShouldContainSubstring, `"instance"`)
		So(string(m), ShouldContainSubstring, `"gc.count"`)
	})
}

func BenchmarkRegistry_GetRegistry(b *testing.B) {
	for n := 0; n < b.N; n++ {
                GlobalRegistry.GetRegistry()
        }
}