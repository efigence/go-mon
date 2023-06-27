package mon

import (
	"fmt"
	"net/http"
	"strings"
)

type PrometheusHandler struct {
}

var prometheusTypes = map[string]string{
	MetricTypeGauge:        "gauge",
	MetricTypeGaugeInt:     "gauge",
	MetricTypeCounter:      "counter",
	MetricTypeCounterFloat: "counter",
}
var promRepl = strings.NewReplacer(
	".", ":",
	"-", "_",
)

func HandlePrometheus(w http.ResponseWriter, req *http.Request) {
	for k, metric := range GlobalRegistry.GetRegistry().Metrics {
		k = promRepl.Replace(k)
		fmt.Fprintf(w, "# HELP %s\n", k)
		if len(metric.Type()) > 0 {
			fmt.Fprintf(w, "# TYPE %s %s\n", k, prometheusTypes[metric.Type()])
		}
		if len(metric.Unit()) > 0 {
			fmt.Fprintf(w, "# UNIT %s %s\n", k, metric.Unit())
		}
		fmt.Fprintf(w, "%s %f\n", k, metric.Value())
		fmt.Fprintf(w, "\n")
	}
}
