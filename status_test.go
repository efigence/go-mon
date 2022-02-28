package mon

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSummaryMessage(t *testing.T) {
	s := NewStatus("testStatus")
	c1, _ := s.NewComponent("db")
	c2, _ := s.NewComponent("web")
	c3, _ := s.NewComponent("batch")
	c4, _ := s.NewComponent("backup")
	c5, _ := s.NewComponent("transcoder")

	assert.NoError(t, c1.Update(StateOk, "msg1"))
	assert.NoError(t, c2.Update(StateWarning, "msg2"))
	assert.NoError(t, c3.Update(StateCritical, "msg3"))
	assert.NoError(t, c4.Update(StateUnknown, "msg4"))
	assert.Error(t, c5.Update(123, "msg5"))

	assert.Contains(t, s.GetMessage(), "msg1")
	assert.Contains(t, s.GetMessage(), "msg2")
	assert.Contains(t, s.GetMessage(), "msg3")
	assert.Contains(t, s.GetMessage(), "msg4")
	assert.NotContains(t, s.GetMessage(), "msg5")
	assert.Equal(t, s.GetState(), StateCritical)

}

func TestSummaryState(t *testing.T) {
	s := NewStatus("testStatus")
	c1, _ := s.NewComponent("db")
	t.Run("After-init state unknown", func(t *testing.T) {
		assert.NotEqual(t, s.GetState(), StateOk)
		assert.Equal(t, s.GetState(), StateUnknown)
		assert.NotEqual(t, c1.GetState(), StateOk)
		assert.Equal(t, c1.GetState(), StateUnknown)
	})
	c2, _ := s.NewComponent("storage")
	c3, _ := s.NewComponent("kk")
	c2.Update(StateCritical, "bad things happened")
	c3.Update(StateOk, "bad things happened")
	t.Run("Should pick most dangerous state of all subservices", func(t *testing.T) {
		assert.Equal(t, s.GetState(), StateCritical)
		assert.Equal(t, c2.GetState(), StateCritical)
		assert.Equal(t, c3.GetState(), StateOk)
	})
}

func TestCreation(t *testing.T) {
	s := NewStatus("testStatus", "with long name", "and description")
	_, err1 := s.NewComponent("db")
	assert.Nil(t, err1)
	_, err2 := s.NewComponent("db")
	assert.Error(t, err2, "do not allow double create")
}
func TestBadInput(t *testing.T) {
	s := NewStatus("testStatus", "with long name", "and description")
	c1 := s.MustNewComponent("db", "some db")
	err3 := s.Update(StateOk, "some message")
	assert.Error(t, err3, "do not allow updating parent that has children")
	err4 := c1.Update(234, "badState")
	assert.Error(t, err4, "Do not allow updating with state code out of range")
	assert.Panics(t, func() { s.MustUpdate(222, "test") })
	assert.Panics(t, func() { s.MustNewComponent("db") })

}

func TestFormatters(t *testing.T) {
	s := NewStatus("testStatus", "with long name", "and description")
	c1, _ := s.NewComponent("db", "some db")
	c1.Update(StateCritical, "some message")
	_ = c1
	out := s.GetMessage()
	assert.NotContains(t, out, "=#=")
}
