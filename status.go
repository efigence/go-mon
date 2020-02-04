package mon

import (
	"fmt"
	"strings"
	"sync"
	"time"
)


type State uint8

const StateInvalid = State(0)
const StateOk = State(1)
const StateWarning = State(2)
const StateCritical = State(3)
const StateUnknown = State(4)
const stateEnd = State(5)

const Invalid = State(0)
const Ok = State(1)
const Warning = State(2)
const Critical = State(3)
const Unknown = State(4)


// Status forms hierarchical structure. Parent status code and message is always generated from status of children so running update on it is pointless
type Status struct {
	State State `json:"state"`
	// Canonical service name (required)
	Name string `json:"name"`
	// FQDN
	FQDN string `json:"fqdn"`
	// Pretty display name of service
	DisplayName string `json:"display_name,nonempty"`
	// Description of service
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
	summaryState func(*map[string]*Status)(state State)
	summaryMessage func(*map[string]*Status)(message string)
	sync.RWMutex
}




// NewStatus creates new status object with state set to unknown
// optional parameters are
// * display name
// * description
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
// Update updates state of the Status component. It should be used only on component with no children or else it will err out
func (s *Status)Update(status State, message string) error {
	s.Lock()
	defer s.Unlock()

	if status > stateEnd || status < 0 {
		return fmt.Errorf("status[%d] outside of range", status)
	}
	if len(s.Components) > 0 {
		return fmt.Errorf("status[%s] have %d children nodes[], updating parent is pointless",s.Name,len(s.Components))
	}
	s.State = State(status)
	s.Msg = message
	if s.State == StateOk { s.Ok = true }
	s.Ts = time.Now()
	return nil
}

// MustUpdate runs Update and panics on error
//

func (s *Status)MustUpdate(status State, message string) {
	err := s.Update(status,message)
	if err != nil {
		panic(fmt.Sprintf("updating component %s failed: %s", s.Name,err))
	}

}

// NewComponent adds a new child component to the Status
// optional parameters are
// * display name
// * description
func (s *Status)NewComponent(name string, p ...string) (*Status, error) {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.Components[name]; ok {
		return nil, fmt.Errorf("Given component already exists!")
	}
	s.Components[name] = NewStatus(name, p...)
	return s.Components[name], nil
}

func (s *Status)MustNewComponent(name string, p ... string) (*Status) {
	c, err := s.NewComponent(name, p...)
	if err != nil {
		panic(fmt.Sprintf("error when creating new component %s: %s",name, err))
	}
	return c
}
// update and return message
func (s *Status)GetMessage() string{
	s.RLock()
	defer s.RUnlock()

	if len(s.Components) > 0 {
		s.Msg =  s.summaryMessage(&s.Components)
		// FIXME that should be called in reverse relation, child updating parent
		for _, v :=  range s.Components {
			if v.Ts.After(s.Ts) {
				s.Ts = v.Ts
			}
		}
	}
	return s.Msg
}

// update and return message
func (s *Status)GetState() State{
	s.RLock()
	defer s.RUnlock()

	if len(s.Components) > 0 {
		s.State =  s.summaryState(&s.Components)
	}
	return s.State
}

// SummarizeStatusState returns highest ( critical>unknown>warning>ok ) state of underlying status map
func SummarizeStatusState(component *map[string]*Status)(state State) {
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
