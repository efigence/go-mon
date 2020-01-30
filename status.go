package mon

import (
	"fmt"
	"strings"
	"time"
	"math"
	"sync"
)


const StateInvalid = 0
const StateOk = 1
const StateWarning = 2
const StateCritical = 3
const StateUnknown = 4

const Invalid = 0
const Ok = 1
const Warning = 2
const Critical = 3
const Unknown = 4


// Status forms hierarchical structure. Parent status code and message is always generated from status of children so running update on it is pointless
type Status struct {
	State uint8 `json:"state"`
	// Canonical service name (required)
	Name string `json:"name"`
	// FQDN
	FQDN string `json:"fqdn"`
	// Pretty display name of service
	DisplayName string `json:"display_name,nonempty"`
	// Description of serive
	Description string `json:"description,nonempty"`
	// status check message
	Msg string `json:"msg"`
	// data format initialization canary.
	// Proper implementation will set ok to true if status is really okay
	// but fresh (all fields zero) object will be invalid (state = 0 but ok = false)
	// and that can be detected upstream.
	// Other function is to allow just checking one bool flag to decide if it is ok or not
	Ok bool `json:"ok"`
	Ts time.Time `json:"ts"`
	Components map[string]*Status `json:"components,nonempty"`
	// function used to generate status and message from underlying components
	summaryState func(*map[string]*Status)(state uint8)
	summaryMessage func(*map[string]*Status)(message string)
	sync.RWMutex
}




// NewStatus creates new status object with state set to unknown
func NewStatus(name string, p ...string) *Status {
	var s Status
	s.Name = name
	if len(p) > 0 { s.DisplayName = p[0] }
	if len(p) > 1 { s.Description = p[1] }
	s.Components = make(map[string]*Status)
	s.State = StateUnknown
	s.Ok = false
	s.summaryMessage = SummarizeStatusMessage
	s.summaryState = SummarizeStatusState

	return &s
}

func (s *Status)Update(status int, message string) error {
	s.Lock()
	defer s.Unlock()

	if status > math.MaxUint8 || status < 0 {
		return fmt.Errorf("status[%d] outside of range", status)
	}
	if len(s.Components) > 0 {
		return fmt.Errorf("status[%s] have %d children nodes[], updating parent is pointless",s.Name,len(s.Components))
	}
	s.State = uint8(status)
	s.Msg = message
	if s.State == StateOk { s.Ok = true }
	return nil
}


func (s *Status)NewComponent(name string, p ...string) (*Status, error) {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.Components[name]; ok {
		return nil, fmt.Errorf("Given component already exists!")
	}
	s.Components[name] = NewStatus(name, p...)
	return s.Components[name], nil
}
// update and return message
func (s *Status)GetMessage() string{
	s.RLock()
	defer s.RUnlock()

	if len(s.Components) > 0 {
		s.Msg =  s.summaryMessage(&s.Components)
	}
	return s.Msg
}

// update and return message
func (s *Status)GetState() uint8{
	s.RLock()
	defer s.RUnlock()

	if len(s.Components) > 0 {
		s.State =  s.summaryState(&s.Components)
	}
	return s.State
}

// SummarizeStatusState returns highest ( critical>unknown>warning>ok ) state of underlying status map
func SummarizeStatusState(component *map[string]*Status)(state uint8) {
	for _, c := range *component {
		switch {
		// Critical state is always most important one to report; nothing to do after if we find one
		case c.State == StateCritical:
			return StateCritical
		case c.State > state:
			state = c.State
		}
	}
	return state
}
// SummarizeStatusMessage generates status message based on map of components and their statuses
func SummarizeStatusMessage(component *map[string]*Status)(message string) {
	var sCritical, sWarning ,sUnknown, sOk []string
	for _, c := range *component {
		componentInfo := fmt.Sprintf("[%s]%s",c.Name,c.Msg)
		switch c.State {
		case StateOk: sOk = append(sOk, componentInfo)
		case StateWarning: sWarning = append(sWarning, componentInfo)
		case StateCritical: sCritical = append(sCritical, componentInfo)
		default: sUnknown = append(sUnknown, componentInfo)
		}

	}
	var outArr []string
	if len(sCritical) > 0 { outArr = append(outArr,"C:" + strings.Join(sCritical,", ")) }
	if len(sWarning) > 0  { outArr = append(outArr,"W:" + strings.Join(sWarning,", ")) }
	if len(sUnknown) > 0  { outArr = append(outArr,"U:" + strings.Join(sUnknown,", ")) }
	if len(sOk) > 0       { outArr = append(outArr,       strings.Join(sOk,", "))}
	return strings.Join(outArr, " -=#=- ")
}
