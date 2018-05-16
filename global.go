package mon

import (
	"path/filepath"
	"os"
	"fmt"
)

// Global registry, will use app's executable name as instance and try best to guess FQDN
// You can change thos via Set..() family of methods
var GlobalRegistry *Registry
var GlobalStatus *Status
func init() {
	_, name := filepath.Split(os.Args[0])
	r, err := NewRegistry(getFQDN(),name,10)
	if err != nil {
		panic(fmt.Sprintf("could not create global registry: %s",err))
	}
	GlobalRegistry = r
	GlobalStatus = NewStatus(name)
}

