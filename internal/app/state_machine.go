package app

import "fmt"

type State string

const (
	BootMenuState     State = "boot_menu"
	NewGameFormState  State = "new_game_form"
	LoadGameListState State = "load_game_list"
	HelpState         State = "help"
	LoadingGameState  State = "loading_game"
	DashboardState    State = "dashboard"
	FatalErrorState   State = "fatal_error"
)

type Event string

const (
	EventGoNewGame      Event = "go_new_game"
	EventGoLoadGame     Event = "go_load_game"
	EventGoHelp         Event = "go_help"
	EventBack           Event = "back"
	EventExit           Event = "exit"
	EventSubmitNewGame  Event = "submit_new_game"
	EventSelectSave     Event = "select_save"
	EventLoadSucceeded  Event = "load_succeeded"
	EventLoadFailed     Event = "load_failed"
	EventRecoverToBoot  Event = "recover_to_boot"
	EventRecoverToLoad  Event = "recover_to_load"
	EventShowFatalError Event = "show_fatal_error"
)

type Machine struct {
	state State
}

func NewMachine() *Machine {
	return &Machine{state: BootMenuState}
}

func (m *Machine) State() State {
	return m.state
}

func (m *Machine) Transition(e Event) error {
	next, ok := transition(m.state, e)
	if !ok {
		return fmt.Errorf("invalid transition: state=%s event=%s", m.state, e)
	}
	m.state = next
	return nil
}

func transition(s State, e Event) (State, bool) {
	switch s {
	case BootMenuState:
		switch e {
		case EventGoNewGame:
			return NewGameFormState, true
		case EventGoLoadGame:
			return LoadGameListState, true
		case EventGoHelp:
			return HelpState, true
		case EventExit:
			return BootMenuState, true
		case EventShowFatalError:
			return FatalErrorState, true
		}
	case NewGameFormState:
		switch e {
		case EventSubmitNewGame:
			return LoadingGameState, true
		case EventBack:
			return BootMenuState, true
		case EventShowFatalError:
			return FatalErrorState, true
		}
	case LoadGameListState:
		switch e {
		case EventSelectSave:
			return LoadingGameState, true
		case EventBack:
			return BootMenuState, true
		case EventGoNewGame:
			return NewGameFormState, true
		case EventShowFatalError:
			return FatalErrorState, true
		}
	case HelpState:
		switch e {
		case EventBack:
			return BootMenuState, true
		case EventShowFatalError:
			return FatalErrorState, true
		}
	case LoadingGameState:
		switch e {
		case EventLoadSucceeded:
			return DashboardState, true
		case EventRecoverToBoot:
			return BootMenuState, true
		case EventRecoverToLoad:
			return LoadGameListState, true
		case EventShowFatalError:
			return FatalErrorState, true
		}
	case DashboardState:
		switch e {
		case EventBack:
			return BootMenuState, true
		case EventShowFatalError:
			return FatalErrorState, true
		}
	case FatalErrorState:
		switch e {
		case EventRecoverToBoot:
			return BootMenuState, true
		case EventExit:
			return FatalErrorState, true
		}
	}
	return "", false
}
