package mon

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// HandleMetrics is basic web hook that returns JSON dump of metrics in GlobalRegistry
func HandleMetrics( w http.ResponseWriter, req *http.Request) {
	js, err := json.Marshal(GlobalRegistry.GetRegistry())
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.Write([]byte(`{"msg":"JSON marshalling error"}`))
	} else {
		w.Write(js)
	}

}
// HandleHealthchecks returns GlobalStatus with appropriate HTTP code
func HandleHealthcheck ( w http.ResponseWriter, req *http.Request) {
	var httpStatus int
	w.Header().Set("Content-Type", "application/json")
	switch GlobalStatus.GetState() {
	case StateOk:
		httpStatus =  http.StatusOK
	case StateWarning:
		httpStatus =  http.StatusOK
	case StateUnknown:
		httpStatus =  http.StatusInternalServerError
	case StateInvalid:
		httpStatus =  http.StatusInternalServerError
	default:
		httpStatus =  http.StatusServiceUnavailable
	}
	js, err := json.Marshal(GlobalStatus)
	if err != nil {
		http.Error(w, err.Error(), httpStatus)
		return
	} else if httpStatus != http.StatusOK {
		w.WriteHeader(httpStatus)
	}
	w.Write(js)
}



type HaproxyState struct {
	State State
	BackendName string
	ServerName string
	LBNodeName string
	ServerWeight int
	TotalWeight int
	// Current connections going to this server
	ServerCurrentConnections int
	// Current connections going to backend
	BackendCurrentConnections int
	// Requests in server queue
	Queue int
	// whether header was found
	Found bool
	sync.RWMutex
}
// SafeToStop returns whether it is safe to shutdown the server.
// it will only return true if in status:
//
// * server is not in UP state
// * there is no active or queued connections to it
// * there is no haproxy server state header present

func (s *HaproxyState) SafeToStop() bool {
	s.RLock()
	defer s.RUnlock()
	if s.State == Ok {return false}
	if s.State == StateInvalid { return true}
	if s.Queue == 0 && s.ServerCurrentConnections == 0 {
		return true
	}
	return false
}

func (s *HaproxyState) copy(new HaproxyState) {
	s.Lock()
	defer s.Unlock()
	s.State = new.State
	s.BackendName = new.BackendName
	s.ServerName = new.ServerName
	s.LBNodeName = new.LBNodeName
	s.ServerWeight = new.ServerWeight
	s.TotalWeight = new.TotalWeight
	s.ServerCurrentConnections = new.ServerCurrentConnections
	s.BackendCurrentConnections = new.BackendCurrentConnections
	s.Found = new.Found
	s.Queue = new.Queue
}
// HandleHaproxyState parses haproxy state header and returns current backend state
//
// Example header:  X-Haproxy-Server-State: UP 2/3; name=bck/srv2; node=lb1; weight=1/2; scur=13/22; qcur=
func HandleHaproxyState ( req *http.Request) (haproxyState HaproxyState, found bool, err error) {
	var s HaproxyState
	stateSlice := strings.Split(req.Header.Get("X-Haproxy-Server-State"), ";")
	if len(stateSlice) < 2 { return }
	if strings.Contains(stateSlice[0],"UP") {
		s.State = Ok
	} else if strings.Contains(stateSlice[0],"DOWN") {
		s.State = Critical
	} else if strings.Contains(stateSlice[0],"NOLB") {
		s.State = Warning
	} else {
		return
	}

	for _, part := range stateSlice[1:] {
		part = strings.TrimSpace(part)
		ss := strings.Split(strings.TrimSpace(part),"=")
		if len(ss) < 2 {continue}
		k := strings.TrimSpace(ss[0])
		v := strings.TrimSpace(ss[1])
		switch k {
		case "name":
			sss :=  strings.SplitN(v,"/",2)
			if len(sss) == 2 {
				s.BackendName = sss[0]
				s.ServerName = sss[1]
			} else {
				s.ServerName = v
			}
		case "node":
			s.LBNodeName=v
		case "weight":
			sss := strings.SplitN(v,"/",2)
			if len(sss) == 2 {
				i1, err1 := strconv.Atoi(sss[0])
				i2, err2 := strconv.Atoi(sss[1])
				if err1 != nil || err2 != nil {
					return s, true, fmt.Errorf("error parsing [%s]: %s|%s", v, err1, err2)
				}
				s.ServerWeight = i1
				s.TotalWeight = i2
			}
		case "scur":
			sss := strings.SplitN(v,"/",2)
			if len(sss) == 2 {
				i1, err1 := strconv.Atoi(sss[0])
				i2, err2 := strconv.Atoi(sss[1])
				if err1 != nil || err2 != nil {
					return s, true, fmt.Errorf("error parsing [%s]: %s|%s", v, err1, err2)
				}
				s.ServerCurrentConnections = i1
				s.BackendCurrentConnections = i2
			}
		case "qcur":
			i, err := strconv.Atoi(v)
			if err != nil {
					return s, true, fmt.Errorf("error parsing [%s]: %s", v, err)
			}
			s.Queue = i
		}
	}
	s.Found = true
	return s, true, nil
}

// HandleHealthchecksHaproxy returns GlobalStatus with appropriate HTTP code and handles X-Haproxy-Server-State header
func HandleHealthchecksHaproxy() (handlerFunc func ( w http.ResponseWriter, req *http.Request), haproxyState *HaproxyState) {
	var state = &HaproxyState{}
	return func ( w http.ResponseWriter, req *http.Request) {
		newState, _, err := HandleHaproxyState(req)
		if err == nil {
			state.copy(newState)
		} else {
			state.copy(HaproxyState{})
		}
		HandleHealthcheck (w,req)
	}, state
}