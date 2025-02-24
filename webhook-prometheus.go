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
	emittedHelp := map[string]bool{}
	for k, m1 := range registry.GetRegistry().Metrics {
		for k2, metric := range m1 {
			k = promRepl.Replace(k)
			t := ""
			keyName := k
			metricType := metric.Type()
			metricUnit := metric.Unit()
			if metricUnit != "" {
				if metricType == MetricTypeCounter || metricType == MetricTypeCounterFloat {
					keyName = keyName + "_" + metricUnit + "_total"
				} else {
					keyName = keyName + "_" + metricUnit
				}
			}

			if k2 != string(emptyGob) {
				tags := ungobTag([]byte(k2))
				tagSlice := goneric.MapToSlice(
					func(k string, v string) string {
						return k + "=" + `"` + v + `"`
					},
					tags.T)
				t = "{" + strings.Join(tagSlice, ",") + "}"
			}
			if _, ok := emittedHelp[keyName]; !ok {
				fmt.Fprintf(w, "\n# HELP %s\n", keyName)
				if len(metric.Type()) > 0 {
					fmt.Fprintf(w, "# TYPE %s %s\n", keyName, prometheusTypes[metric.Type()])
				}
				if len(metric.Unit()) > 0 {
					fmt.Fprintf(w, "# UNIT %s %s\n", keyName, metric.Unit())
				}
				emittedHelp[keyName] = true
			}

			fmt.Fprintf(w, "%s%s %f\n", keyName, t, metric.Value())

		}
	}
}
