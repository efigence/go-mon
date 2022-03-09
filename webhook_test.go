package mon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsHandler(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/metrics", nil)

	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleMetrics)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	assert.Equal(t, http.StatusOK, rr.Code, "status code")
	assert.Contains(t, rr.Body.String(), "gc.heap_idle", "contains data")
}

func TestStatusHandler(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/health", nil)

	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleHealthcheck)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// initial status should be invalid (because nobody updated it)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "no update should result in invalid data")
	var s1 Status
	err = json.Unmarshal(rr.Body.Bytes(), &s1)
	assert.Nil(t, err)
	assert.Equal(t, State(StatusUnknown), s1.State, "unknown if there was no status change from init")

	// change status to OK
	err = GlobalStatus.Update(StatusOk, "service-running")
	require.Nil(t, err)
	req, err = http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(HandleHealthcheck)
	handler.ServeHTTP(rr, req)
	var s2 Status
	err = json.Unmarshal(rr.Body.Bytes(), &s2)
	assert.Equal(t, http.StatusOK, rr.Code, "status ok after setting it")
	assert.Nil(t, err)
	assert.Equal(t, State(StatusOk), s2.State)
	assert.Contains(t, rr.Body.String(), `service-running`, "json returns service is running")

	// change status to critical
	GlobalStatus.Update(StatusCritical, "service-failed")
	req, err = http.NewRequest("GET", "/health", nil)
	require.NoError(t, err)
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(HandleHealthcheck)
	handler.ServeHTTP(rr, req)
	var s3 Status
	err = json.Unmarshal(rr.Body.Bytes(), &s3)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rr.Code, "unavailable status after service is down")
	assert.Equal(t, State(StatusCritical), s3.State, "status critical")
	assert.Contains(t, rr.Body.String(), `service-failed`, "json returns service failed")
}

func TestHandleHaproxyState_Up(t *testing.T) {
	req, err := http.NewRequest("GET", "/metrics", nil)
	require.NoError(t, err)
	req.Header.Add(
		"X-Haproxy-Server-State",
		"UP 2/3; name=bck/srv2; node=lb1; weight=1/2; scur=13/22; qcur=6",
	)
	var state HaproxyState
	var stateErr error
	var stateFound bool
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		state, stateFound, stateErr = HandleHaproxyState(req)
	})
	handler.ServeHTTP(rr, req)

	assert.True(t, stateFound, "haproxy header")
	assert.NoError(t, stateErr)
	assert.Equal(t, "bck", state.BackendName, "backend name")
	assert.Equal(t, "srv2", state.ServerName, "server name")
	assert.Equal(t, "lb1", state.LBNodeName, "LB name")
	assert.Equal(t, 1, state.ServerWeight, "current server weight")
	assert.Equal(t, 2, state.TotalWeight, "backend weight sum")
	assert.Equal(t, 13, state.ServerCurrentConnections, "server connection count")
	assert.Equal(t, 22, state.BackendCurrentConnections, "backend connection count")
	assert.Equal(t, 6, state.Queue, "queue to server")
	assert.False(t, state.SafeToStop(), "not safe to stop")
}
func TestHandleHaproxyState_Down(t *testing.T) {
	req, err := http.NewRequest("GET", "/metrics", nil)
	require.NoError(t, err)
	req.Header.Add(
		"X-Haproxy-Server-State",
		"DOWN 2/3; name=bck/srv2; node=lb1; weight=1/2; scur=0/22; qcur=0",
	)
	var state HaproxyState
	var stateErr error
	var stateFound bool
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		state, stateFound, stateErr = HandleHaproxyState(req)
	})
	handler.ServeHTTP(rr, req)

	assert.True(t, stateFound, "haproxy header")
	assert.NoError(t, stateErr)
	assert.Equal(t, "bck", state.BackendName, "backend name")
	assert.Equal(t, "srv2", state.ServerName, "server name")
	assert.Equal(t, "lb1", state.LBNodeName, "LB name")
	assert.Equal(t, 1, state.ServerWeight, "current server weight")
	assert.Equal(t, 2, state.TotalWeight, "backend weight sum")
	assert.Equal(t, 0, state.ServerCurrentConnections, "server connection count")
	assert.Equal(t, 22, state.BackendCurrentConnections, "backend connection count")
	assert.Equal(t, 0, state.Queue, "queue to server")
	assert.True(t, state.SafeToStop(), "safe to stop")
}

func TestHandleHaproxyState_Empty(t *testing.T) {
	req, err := http.NewRequest("GET", "/metrics", nil)
	require.NoError(t, err)
	var state HaproxyState
	var stateErr error
	var stateFound bool
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		state, stateFound, stateErr = HandleHaproxyState(req)
	})
	handler.ServeHTTP(rr, req)

	assert.False(t, stateFound, "haproxy header")
	assert.NoError(t, stateErr)
	assert.True(t, state.SafeToStop(), "safe to stop")
}

func TestHandleHealthchecksHaproxy(t *testing.T) {
	req, err := http.NewRequest("GET", "/metrics", nil)
	require.NoError(t, err)
	req.Header.Add(
		"X-Haproxy-Server-State",
		"DOWN 2/3; name=bck/srv2; node=lb1; weight=1/2; scur=0/22; qcur=0",
	)
	GlobalStatus.Update(StatusOk, "service-running")
	checkHandler, haproxyStatus := HandleHealthchecksHaproxy(true)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(checkHandler)
	handler.ServeHTTP(rr, req)

	haproxyStatus.RLock()
	defer haproxyStatus.RUnlock()
	assert.True(t, haproxyStatus.Found, "haproxy header")
	assert.Equal(t, "bck", haproxyStatus.BackendName, "backend name")
	assert.Equal(t, "srv2", haproxyStatus.ServerName, "server name")
	assert.Equal(t, "lb1", haproxyStatus.LBNodeName, "LB name")
	assert.Equal(t, 1, haproxyStatus.ServerWeight, "current server weight")
	assert.Equal(t, 2, haproxyStatus.TotalWeight, "backend weight sum")
	assert.Equal(t, 0, haproxyStatus.ServerCurrentConnections, "server connection count")
	assert.Equal(t, 22, haproxyStatus.BackendCurrentConnections, "backend connection count")
	assert.Equal(t, 0, haproxyStatus.Queue, "queue to server")
	assert.True(t, haproxyStatus.SafeToStop(), "safe to stop")
	assert.Equal(t, http.StatusOK, rr.Code, "status OK")
	assert.Contains(t, rr.Body.String(), `"state":1`, "state body")
	assert.Contains(t, rr.Body.String(), "service-running")
}
