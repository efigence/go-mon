package mon

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"os"
)
func TestGetFQDN(t *testing.T) {
	hostname,_ := os.Hostname()
	Convey("ShouldReturnAllMessages", t, func() {
		// TODO that would probably fail on misconfigured machine... not my problem, fix your shit
		So(getFQDN(), ShouldContainSubstring, ".")
		So(getFQDN(), ShouldContainSubstring, hostname)
	})
}