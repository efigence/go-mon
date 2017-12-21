package mon

import (
	"net/http"
	"gopkg.in/unrolled/render.v1"
)
var r = render.New()

// HandleMetrics is basic web hook that returns JSON dump of metrics in GlobalRegistry
func HandleMetrics( w http.ResponseWriter, req *http.Request) {
	r.JSON(w, http.StatusOK,  GlobalRegistry)

}

func HandleHealthcheck ( w http.ResponseWriter, req *http.Request) {
	switch GlobalStatus.State {
	case StateOk:
		r.JSON(w, http.StatusOK,  GlobalStatus)
	case StateWarning:
		r.JSON(w, http.StatusOK,  GlobalStatus)
	case StateUnknown:
		r.JSON(w, http.StatusInternalServerError,  GlobalStatus)
	case StateInvalid:
		r.JSON(w, http.StatusInternalServerError,  GlobalStatus)
	default:
		r.JSON(w, http.StatusServiceUnavailable,  GlobalStatus)
	}

}
