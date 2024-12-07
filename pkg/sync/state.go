package sync

import (
	"context"
	"errors"
	"fmt"
	acerrors "github.com/zhughes3/go-accelerate/pkg/errors"
	"github.com/zhughes3/go-accelerate/pkg/slog"
	"sync"
)

// State represents the operational status of a component.
// The components can be in one of the following states:
//
//		New, when the component has been newly created.
//	 Started, when the component is actively running or processing.
//	 Shutdown, when the component has been stopped or terminated.
type State int

const (
	StateNew = iota
	StateStarted
	StateShutdown
)

func (s State) String() string {
	switch s {
	case StateNew:
		return "New"
	case StateStarted:
		return "Started"
	case StateShutdown:
		return "Shutdown"
	default:
		return fmt.Sprintf("%d is not a valid date", s)
	}
}

type StateMachineBuilder struct {
	holder StateMachine
}

func NewStateMachineBuilder(logger slog.Logger) *StateMachineBuilder {
	return &StateMachineBuilder{holder: StateMachine{
		logger:        logger,
		componentName: "",
		mu:            sync.Mutex{},
		current:       StateNew,
	}}
}

func (s *StateMachineBuilder) WithComponentName(name string) *StateMachineBuilder {
	s.holder.componentName = name
	return s
}

func (s *StateMachineBuilder) WithIgnoreAlreadyAtEndError(b bool) *StateMachineBuilder {
	s.holder.ignoreAlreadyAtEnd = b
	return s
}

func (s *StateMachineBuilder) Build() *StateMachine {
	return &s.holder
}

type StateMachine struct {
	logger slog.Logger

	componentName string

	// ignoreAlreadyAtEnd should be true if a user wants to "ignore" duplicate requests to move to an end state if
	// the current is already at that state. For example, if someone calls [StateMachine.Shutdown] twice, the 2nd time
	// will not error if this value is true.
	ignoreAlreadyAtEnd bool

	// mu is used to guard the [current] state since it is mutable
	mu sync.Mutex

	// current holds the current state of a component
	current State
}

// errAlreadyAtEnd is a sentinel error that is returned from [StateMachine.Transition] to indicate the current state
// is already in the wanted state
var errAlreadyAtEnd = errors.New("state is already at wanted state")

func (s *StateMachine) Start(ctx context.Context, fn func() error) error {
	err := s.doTransition(StateNew, StateStarted, fn)

	return s.stateErrorIfNeeded(ctx, err, startVerbs)
}

func (s *StateMachine) Shutdown(ctx context.Context, fn func() error) error {
	err := s.doTransition(StateStarted, StateShutdown, fn)

	return s.stateErrorIfNeeded(ctx, err, shutdownVerbs)
}

func (s *StateMachine) doTransition(expected State, wanted State, fn func() error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == wanted {
		return errAlreadyAtEnd
	}
	if s.current != expected {
		return fmt.Errorf("cannot transition to %v because the current state is %v, not %v", wanted, s.current, expected)
	}

	if err := fn(); err != nil {
		return err
	}

	s.current = wanted

	return nil
}

func (s *StateMachine) stateErrorIfNeeded(ctx context.Context, err error, verbs operationVerbs) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, errAlreadyAtEnd) && s.ignoreAlreadyAtEnd {
		// TODO log
		//fmt.Printf("%s already %s. Ignoring subsequent %s", s.componentName, verbs.PastTense, verbs.PresentTense)
		return nil
	}

	return acerrors.Wrapf(err, "problem %s %s", verbs.presentParticiple, s.componentName)
}

type operationVerbs struct {
	presentTense      string
	pastTense         string
	presentParticiple string
}

var (
	shutdownVerbs = operationVerbs{
		presentTense:      "shutdown",
		pastTense:         "shutdown",
		presentParticiple: "shutting down",
	}
	startVerbs = operationVerbs{
		presentTense:      "start",
		pastTense:         "started",
		presentParticiple: "starting",
	}
)
