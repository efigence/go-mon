package mon

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSummaryMessage(t *testing.T) {
	s := NewStatus("testStatus")
	c1, _ := s.NewComponent("db")
	c2, _ := s.NewComponent("web")
	c3, _ := s.NewComponent("batch")
	c4, _ := s.NewComponent("backup")
	c5, _ := s.NewComponent("transcoder")

	require.NoError(t, c1.Update(StateOk, "msg1"))
	require.NoError(t, c2.Update(StateWarning, "msg2"))
	require.NoError(t, c3.Update(StateCritical, "msg3"))
	require.NoError(t, c4.Update(StateUnknown, "msg4"))
	assert.Error(t, c5.Update(123, "msg5"))
	assert.Contains(t, s.GetMessage(), "msg1", "<%s> - <%s>", s.Msg, c1.Msg)
	assert.Contains(t, s.GetMessage(), "msg2", "<%s> - <%s>", s.Msg, c2.Msg)
	assert.Contains(t, s.GetMessage(), "msg3", "<%s> - <%s>", s.Msg, c3.Msg)
	assert.Contains(t, s.GetMessage(), "msg4", "<%s> - <%s>", s.Msg, c4.Msg)
	assert.NotContains(t, s.GetMessage(), "msg5")
	assert.Equal(t, s.GetState(), StateCritical, "state: %+v", s)

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
	c1, err1 := s.NewComponent("db")
	assert.Nil(t, err1)
	assert.False(t, s.Ok, "status should not be okay after creation")
	assert.Nil(t, c1.Update(StateOk, "test"))
	assert.Equal(t, s.GetState(), StateOk)
	assert.True(t, s.GetOK(), "ok should flip to true after setting state to OK: %+v", s)
	_, err2 := s.NewComponent("db")
	assert.Error(t, err2, "do not allow double create")
}

func TestInhertitance(t *testing.T) {
	s := NewStatus("testStatus")
	c1, err1 := s.NewComponent("s2")
	require.Nil(t, err1)
	t.Run("state unknnown", func(t *testing.T) {
		assert.False(t, s.Ok, "parent")
		assert.Equal(t, s.GetState(), StateUnknown, "parent")
		assert.False(t, c1.Ok, "child")
		assert.Equal(t, c1.GetState(), StateUnknown, "child")
	})
	t.Run("state ok", func(t *testing.T) {
		c1.Update(StateOk, "ok")
		assert.True(t, s.GetOK(), "parent")
		assert.Equal(t, StateOk, s.GetState(), "parent")
		assert.True(t, c1.Ok, "child")
		assert.Equal(t, StateOk, c1.GetState(), "child")
	})
	t.Run("state warning", func(t *testing.T) {
		c1.Update(StateWarning, "bad")
		assert.False(t, s.GetOK(), "parent")
		assert.Equal(t, s.GetState(), StateWarning, "parent")
		assert.False(t, c1.Ok, "child")
		assert.Equal(t, c1.GetState(), StateWarning, "child")
	})

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

func TestSummarizeStatusMessage(t *testing.T) {
	st_ok := NewStatus("c-ok")
	require.Nil(t, st_ok.Update(StateOk, "ok"))
	st_warn := NewStatus("c-warn")
	require.Nil(t, st_warn.Update(StateWarning, "warning"))
	st_crit := NewStatus("c-crit")
	require.Nil(t, st_crit.Update(StateCritical, "critical"))
	st_unknown := NewStatus("c-unknown")
	require.Nil(t, st_unknown.Update(StateUnknown, "unknown"))

	t1 := make(map[string]*Status, 0)
	t1["ok"] = st_ok
	t1["warn"] = st_warn

	t1_out := SummarizeStatusMessage(&t1)
	assert.Contains(t, t1_out, "ok")
	assert.Contains(t, t1_out, "warning")

	t2 := make(map[string]*Status, 0)
	t2["ok"] = st_ok
	t2["warn"] = st_warn
	t2["crit"] = st_crit
	t2["unk"] = st_unknown

	t2_out := SummarizeStatusMessage(&t2)
	assert.Contains(t, t2_out, "ok")
	assert.Contains(t, t2_out, "warning")
	assert.Contains(t, t2_out, "critical")
	assert.Contains(t, t2_out, "unknown")
}

func TestSummarizeStatusState(t *testing.T) {
	st_ok := NewStatus("c-ok")
	require.Nil(t, st_ok.Update(StateOk, "ok"))
	st_warn := NewStatus("c-warn")
	require.Nil(t, st_warn.Update(StateWarning, "warning"))
	st_crit := NewStatus("c-crit")
	require.Nil(t, st_crit.Update(StateCritical, "critical"))
	st_unknown := NewStatus("c-unknown")
	require.Nil(t, st_unknown.Update(StateUnknown, "unknown"))

	t1 := make(map[string]*Status, 0)
	t1["ok"] = st_ok
	t1["warn"] = st_warn

	t1_out := SummarizeStatusState(&t1)
	assert.Equal(t, t1_out, StateWarning)

	t2 := make(map[string]*Status, 0)
	t2["ok"] = st_ok
	t2["warn"] = st_warn
	t2["crit"] = st_crit
	t2["unk"] = st_unknown

	t2_out := SummarizeStatusState(&t2)
	assert.Equal(t, t2_out, StateCritical)

	t3 := make(map[string]*Status, 0)
	t3["ok"] = st_ok
	t3["unk"] = st_unknown

	t3_out := SummarizeStatusState(&t3)
	assert.Equal(t, t3_out, StateUnknown)

}

func TestNestedChecks(t *testing.T) {
	s := NewStatus("testStatus")
	c1 := s.MustNewComponent("db")
	c2 := s.MustNewComponent("web")
	require.NoError(t, c1.Update(StateOk, "msg1"))
	c22 := c2.MustNewComponent("app1")
	assert.Nil(t, c22.Update(StateWarning, "webappwarn2"))
	// tests are async updated, no real better way to fix it
	time.Sleep(time.Millisecond)
	assert.Contains(t, c1.GetMessage(), "msg1")
	assert.Contains(t, c2.GetMessage(), "webapp")
	assert.Contains(t, c2.GetMessage(), "webappwarn2")
	assert.Contains(t, s.GetMessage(), "webapp")
	assert.Contains(t, s.GetMessage(), "webappwarn2")
	assert.Equal(t, StateOk, c1.GetState())
	assert.Equal(t, StateWarning, c2.GetState())
	assert.Equal(t, StateWarning, s.GetState())

}
