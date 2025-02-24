package mon

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlePrometheus(t *testing.T) {
	metric, err := GlobalRegistry.RegisterOrGet("promtest-name.list", NewGauge("cake"))
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
	assert.Contains(t, rr.Body.String(), "# TYPE promtest_name:list_cake gauge\n")
	assert.Contains(t, rr.Body.String(), "# UNIT promtest_name:list_cake cake\n")
	assert.Contains(t, rr.Body.String(), "promtest_name:list_cake 10.")

}
func TestHandlePrometheusTags(t *testing.T) {
	r, err := NewRegistry("", "", 10)
	require.NoError(t, err)
	metric, err := r.RegisterOrGet("promtest-name_list", NewGauge("cake"), map[string]string{"k1": "v1", "k2": "v2"})
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
	handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		handlePrometheus(writer, request, r)
	})

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	assert.Equal(t, http.StatusOK, rr.Code, "status code")
	assert.Contains(t, rr.Body.String(), "# TYPE promtest_name_list_cake gauge\n")
	assert.Contains(t, rr.Body.String(), "# UNIT promtest_name_list_cake cake\n")
	assert.Contains(t, rr.Body.String(), `promtest_name_list`)
	assert.Contains(t, rr.Body.String(), `k1="v1"`)
	assert.Contains(t, rr.Body.String(), `k2="v2"`)
	assert.Contains(t, rr.Body.String(), ` 10.`)

}
