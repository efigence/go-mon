package mon

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestSummaryMessage(t *testing.T) {
	s := NewStatus("testStatus")
	c1, _ := s.NewComponent("db")
	c2, _ := s.NewComponent("web")
	c3, _ := s.NewComponent("batch")
	c4, _ := s.NewComponent("backup")
	c5, _ := s.NewComponent("transcoder")


	c1.Update(StateOk, "msg1")
	c2.Update(StateWarning, "msg2")
	c3.Update(StateCritical, "msg3")
	c4.Update(StateUnknown, "msg4")
	c5.Update(123, "msg5")

	Convey("ShouldReturnAllMessages", t, func() {
		So(s.GetMessage(), ShouldContainSubstring, "msg1")
		So(s.GetMessage(), ShouldContainSubstring, "msg2")
		So(s.GetMessage(), ShouldContainSubstring, "msg3")
		So(s.GetMessage(), ShouldContainSubstring, "msg4")
		So(s.GetMessage(), ShouldContainSubstring, "msg5")
		So(s.GetState(), ShouldEqual,StateCritical)
	})
}

func TestSummaryState(t *testing.T) {
	s := NewStatus("testStatus")
	c1, _ := s.NewComponent("db")
	Convey("After-init state unknown", t, func() {
		So(s.GetState(),ShouldNotEqual,StateOk)
		So(s.GetState(),ShouldEqual,StateUnknown)
		So(c1.GetState(),ShouldNotEqual,StateOk)
		So(c1.GetState(),ShouldEqual,StateUnknown)
	})
	c2, _ := s.NewComponent("storage")
	c3, _ := s.NewComponent("kk")
	c2.Update(StateCritical, "bad things happened")
	c3.Update(StateOk, "bad things happened")
	Convey("Should pick most dangerous state of all subservices", t, func() {
		So(s.GetState(),ShouldEqual,StateCritical)
		So(c2.GetState(),ShouldEqual,StateCritical)
		So(c3.GetState(),ShouldEqual,StateOk)
	})
}

func TestCreation(t *testing.T) {
	s := NewStatus("testStatus","with long name","and description")
	_, err1 := s.NewComponent("db")
	Convey("Create status with component",t,func() {
		So(err1,ShouldBeNil)
	})
	_, err2 := s.NewComponent("db")
	Convey("do not allow double create",t,func() {
		So(err2,ShouldNotBeNil)
	})
}
func TestBadInput(t *testing.T) {
	s := NewStatus("testStatus","with long name","and description")
	c1, _ := s.NewComponent("db","some db")
	err3 := s.Update(StateOk, "some message")
	Convey("Do not allow updating status with children",t,func() {
		So(err3,ShouldNotBeNil)
	})
	err4 := c1.Update(666,"badState")
	Convey("Do not allow updating with state code out of range",t,func() {
		So(err4,ShouldNotBeNil)
	})
}

func TestFormatters(t *testing.T) {
	s := NewStatus("testStatus","with long name","and description")
	c1, _ := s.NewComponent("db","some db")
	c1.Update(StateCritical, "some message")
	_ = c1
	out := s.GetMessage()
	Convey("no separators between empty types",t,func() {
		So(out,ShouldNotContainSubstring,"=#=")
	})

}
