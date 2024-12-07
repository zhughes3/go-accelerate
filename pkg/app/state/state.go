package state

import (
	"fmt"
	"github.com/zhughes3/go-accelerate/pkg/maps"
)

type State int

const (
	// New is a new app
	New State = iota
	// Starting means the app is attempting to start
	Starting
	// Ready means the app is ready
	Ready
	// Stopping means the app is attempting to stop
	Stopping
	// Stopped means the app has stopped
	Stopped
	// Error means the app is in an error state
	Error
)

var (
	stateToName = map[State]string{
		New:      "NEW",
		Starting: "STARTING",
		Ready:    "READY",
		Stopping: "STOPPING",
		Stopped:  "STOPPED",
		Error:    "ERROR",
	}
	nameToState = maps.Inverse(stateToName)
)

func (s State) String() string {
	if name, ok := stateToName[s]; ok {
		return name
	}

	return fmt.Sprintf("unknown State '%d'", s)
}

func (s State) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText implements [encoding.TextUnmarshaler]. This is the complement of [MarshalText] and supports
// turning the string representation of a State back into the integer enum.
// NOTE: this method has to have a pointer receiver since it will change the value it points to after a successful unmarshal.
func (s *State) UnmarshalText(text []byte) error {
	state, err := parseState(string(text))
	if err != nil {
		return err
	}

	*s = state
	return nil
}

func parseState(raw string) (State, error) {
	if state, ok := nameToState[raw]; ok {
		return state, nil
	}

	return 0, fmt.Errorf("'%s is not a valid state", raw)
}
