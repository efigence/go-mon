package mon

import (
	"path/filepath"
	"os"
	"fmt"
)

var GlobalRegistry *Registry
var GlobalStatus *Status
func init() {
	_, name := filepath.Split(os.Args[0])
	hostname, _ := os.Hostname()
	r, err := NewRegistry(hostname,name)
	if err != nil {
		panic(fmt.Sprintf("could not create global registry: %s",err))
	}
	GlobalRegistry = r
	GlobalStatus = NewStatus(name)
}
