package mon

import (
	"fmt"
	"github.com/XANi/goneric"
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
	handlePrometheus(w, req, GlobalRegistry)
}

func handlePrometheus(w http.ResponseWriter, req *http.Request, registry *Registry) {
	for k, m1 := range registry.GetRegistry().Metrics {
		for k2, metric := range m1 {
			k = promRepl.Replace(k)
			fmt.Fprintf(w, "# HELP %s\n", k)
			if len(metric.Type()) > 0 {
				fmt.Fprintf(w, "# TYPE %s %s\n", k, prometheusTypes[metric.Type()])
			}
			if len(metric.Unit()) > 0 {
				fmt.Fprintf(w, "# UNIT %s %s\n", k, metric.Unit())
			}
			t := ""
			if k2 != string(emptyGob) {
				tags := ungobTag([]byte(k2))
				tagSlice := goneric.MapToSlice(
					func(k string, v string) string {
						return k + "=" + `"` + v + `"`
					},
					tags.T)
				t = "{" + strings.Join(tagSlice, ",")
			}
			fmt.Fprintf(w, "%s%s %f\n", k, t, metric.Value())
			fmt.Fprintf(w, "\n")

		}
	}
}
