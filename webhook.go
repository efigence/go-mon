package mon

import (
	"encoding/json"
	"net/http"
)

// HandleMetrics is basic web hook that returns JSON dump of metrics in GlobalRegistry
func HandleMetrics( w http.ResponseWriter, req *http.Request) {
	js, err := json.Marshal(GlobalRegistry.GetRegistry())
	if err != nil {
		w.Write([]byte(`{"msg":"JSON marshalling error"}`))
	} else {
		w.Write(js)
	}

}
// HandleHealthchecks returns GlobalStatus with appropriate HTTP code
func HandleHealthcheck ( w http.ResponseWriter, req *http.Request) {
	var httpStatus int
	w.Header().Set("Content-Type", "application/json")
	switch GlobalStatus.State {
	case StateOk:
		httpStatus =  http.StatusOK
	case StateWarning:
		httpStatus =  http.StatusOK
	case StateUnknown:
		httpStatus =  http.StatusInternalServerError
	case StateInvalid:
		httpStatus =  http.StatusInternalServerError
	default:
		httpStatus =  http.StatusServiceUnavailable
	}
	js, err := json.Marshal(GlobalStatus)

	if httpStatus != http.StatusOK {
		http.Error(w, "server error", httpStatus)
	} else if err != nil {
		http.Error(w, err.Error(), httpStatus)
	}
	w.Write(js)


}
