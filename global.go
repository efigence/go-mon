package mon

import (
	"path/filepath"
	"os"
	"fmt"
)

// Global registry, will use app's executable name as instance and try best to guess FQDN
// You can change thos via Set..() family of methods
var GlobalRegistry *Registry

// GlobalStatus is app-wide status initialized from app's binary name and fqdn on start
// most of the struct memebers should be changed at app's start
// if you app has more than one component you should not be changing state of this object but instantiate component like
//
//      dbState, err := mon.GlobalStatus.NewComponent("db")
//      if err != nil { ... } // err generally happens when trying to make same component twice
//      dbState.Update(mon.Ok,"db running")
//
//  note that staleness detection is not handled by the package so any test should be ran in gorouting and with timeouts


var GlobalStatus *Status
func init() {
	_, name := filepath.Split(os.Args[0])
	fqdn := getFQDN()
	r, err := NewRegistry(fqdn,name,10)
	if err != nil {
		panic(fmt.Sprintf("could not create global registry: %s",err))
	}
	GlobalRegistry = r
	GlobalStatus = NewStatus(name)
	GlobalStatus.DisplayName = name
	GlobalStatus.FQDN=fqdn
}

