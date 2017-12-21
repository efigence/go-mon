package mon

import (
	"net/http"
    "net/http/httptest"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
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
	Convey("Status Code", t, func() {
		So(rr.Code, ShouldEqual, http.StatusOK)
	})

	Convey("Output Data", t, func() {
		So(rr.Body.String(), ShouldContainSubstring, "gc.heap_idle")
	})
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
	Convey("Status Code for unknown", t, func() {
		So(rr.Code, ShouldEqual, http.StatusInternalServerError)
	})
	Convey("Output Data for unknown", t, func() {
		So(rr.Body.String(), ShouldContainSubstring, `"state":4`)
	})

	// change status to OK
	GlobalStatus.Update(StatusOk,"service-running")
	req, err = http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(HandleHealthcheck)
	handler.ServeHTTP(rr, req)
	Convey("Status Code for OK", t, func() {
		So(rr.Code, ShouldEqual, http.StatusOK)
	})
	Convey("Output Data for OK", t, func() {
		So(rr.Body.String(), ShouldContainSubstring, `"state":1`)
	})
	Convey("Output message for OK", t, func() {
		So(rr.Body.String(), ShouldContainSubstring, `service-running`)
	})

	// change status to critical
	GlobalStatus.Update(StatusCritical,"service-failed")
	req, err = http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(HandleHealthcheck)
	handler.ServeHTTP(rr, req)
	Convey("Status Code for OK", t, func() {
		So(rr.Code, ShouldEqual, http.StatusServiceUnavailable)
	})
	Convey("Output Data for OK", t, func() {
		So(rr.Body.String(), ShouldContainSubstring, `"state":3`)
	})
	Convey("Output message for OK", t, func() {
		So(rr.Body.String(), ShouldContainSubstring, `service-failed`)
	})

}