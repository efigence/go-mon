package mon

import (
	"path/filepath"
	"os"
)

var GlobalRegistry *Registry
var GlobalStatus *Status
func init() {
	_, name := filepath.Split(os.Args[0])
	GlobalRegistry = &Registry{
		Instance: name,
		Metrics:  make(map[string]Metric),
	}
	GlobalStatus = NewStatus(name)
}
