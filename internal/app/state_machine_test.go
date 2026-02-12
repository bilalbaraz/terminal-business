package app

import "testing"

func TestMachineValidTransitions(t *testing.T) {
	m := NewMachine()
	if m.State() != BootMenuState {
		t.Fatalf("unexpected initial state: %s", m.State())
	}
	cases := []struct {
		event Event
		want  State
	}{
		{EventGoNewGame, NewGameFormState},
		{EventBack, BootMenuState},
		{EventGoLoadGame, LoadGameListState},
		{EventSelectSave, LoadingGameState},
		{EventLoadSucceeded, DashboardState},
		{EventBack, BootMenuState},
		{EventGoHelp, HelpState},
		{EventBack, BootMenuState},
		{EventShowFatalError, FatalErrorState},
		{EventRecoverToBoot, BootMenuState},
	}
	for _, tc := range cases {
		if err := m.Transition(tc.event); err != nil {
			t.Fatalf("transition error: %v", err)
		}
		if m.State() != tc.want {
			t.Fatalf("after %s got %s want %s", tc.event, m.State(), tc.want)
		}
	}
}

func TestMachineInvalidTransition(t *testing.T) {
	m := NewMachine()
	if err := m.Transition(EventLoadSucceeded); err == nil {
		t.Fatal("expected invalid transition error")
	}
	if m.State() != BootMenuState {
		t.Fatalf("state changed on invalid transition: %s", m.State())
	}
}

func TestTransitionExhaustiveStateCases(t *testing.T) {
	states := []State{BootMenuState, NewGameFormState, LoadGameListState, HelpState, LoadingGameState, DashboardState, FatalErrorState}
	events := []Event{
		EventGoNewGame, EventGoLoadGame, EventGoHelp, EventBack, EventExit, EventSubmitNewGame,
		EventSelectSave, EventLoadSucceeded, EventLoadFailed, EventRecoverToBoot, EventRecoverToLoad,
		EventShowFatalError,
	}
	for _, s := range states {
		for _, e := range events {
			next, ok := transition(s, e)
			if ok && next == "" {
				t.Fatalf("state %s event %s returned empty next", s, e)
			}
		}
	}
}
