package mon

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlePrometheus(t *testing.T) {
	metric, err := GlobalRegistry.RegisterOrGet("promtest", NewGauge("cake"))
	require.NoError(t, err)
	metric.Update(10.001)

	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/metrics", nil)

	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandlePrometheus)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	assert.Equal(t, http.StatusOK, rr.Code, "status code")
	assert.Contains(t, rr.Body.String(), "# TYPE promtest gauge\n")
	assert.Contains(t, rr.Body.String(), "# UNIT promtest cake\n")
	assert.Contains(t, rr.Body.String(), "promtest 10.")

}
