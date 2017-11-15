package mon

import (
	"sync"
	"fmt"
)


var Registry map[string]Metric
var registryLock = &sync.RWMutex{}

func init() {
	Registry = make(map[string]Metric)
}

func Register(name string, metric Metric) (Metric, error)  {
	registryLock.Lock()
	defer registryLock.Unlock()
	if r, ok := Registry[name]; ok {
		return r, fmt.Errorf("there is metric already registered under this name")
	}
	Registry[name] = metric
	return metric, nil
}
