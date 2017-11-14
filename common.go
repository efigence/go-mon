package mon

import (
	"time"
)

const (
	// 0 is invalid state because that is initial value of int variable which can mean that someone didn't bother to actually set the state
	StatusInvalid = iota
	// nagios-compatible block
	// to get compatible state take nagios one and add +1
	// Service is ok
	StatusOk
	// service is in warning state
	// should be only used if service is *actually* working but have some problems that need to be resolved
	// like "disk getting full" or "worker queue is 90% busy"
	StatusWarning
	// Service is in critical state and is not performing its function
	StatusCritical
	// check failed to get status of service (for reason other than "service is not working)
	// i.e. check itself crashed before providing any useful information about service
	StatusUnknown
)

const (
	// 0 is invalid state because that is initial value of int variable which can mean that someone didn't bother to actually set the state
	HostInvalid = iota
	// host (as in "unit running service checks") is up
	HostUp
	// host is directly unavailable
	HostDown
	// host is down because its parent is down (it is impossble to check because device that connects to the host is unavailable
	HostUnreachable
)

type Service struct {
	// name of the host/metahost service is running on
	Host string `json:"host"`
	// name of service
	Service string `json:"service"`
	// numeric state
	State uint8 `json:"state"`
	// timestamp of the check
	Timestamp time.Time `json:"ts"`
	// duration since last state change
	StateDuration time.Duration `json:"duration,omitempty"`
	// sub-service state
	// if service (say web app) have multiple internal components (for example DB backend, video transcoder etc) that allows it to send state of them to the upstream without multiplying amount of service checks
	// Note that status of them **HAVE** to be aggregated into parent's State
	Components []Service `json:"components,omitempty"`
}
