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
