package ui

import (
	"errors"
	"testing"
	"time"

	"terminal-business/internal/app"
	"terminal-business/internal/persistence"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPanicModelDefaultBranch(t *testing.T) {
	m := NewPanicModel("x")
	if _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyDown}); cmd != nil {
		t.Fatal("expected nil cmd on non-action key")
	}
}

func TestRootUpdateFallbackAndPanicRecovery(t *testing.T) {
	m := NewModel(nil, fakeClock{now: time.Unix(1, 0).UTC()}, fakeRNG{}, nil)
	m.machine = &app.Machine{}
	if _, cmd := m.Update(struct{}{}); cmd != nil {
		t.Fatal("expected nil cmd")
	}

	m2 := NewModel(nil, fakeClock{now: time.Unix(1, 0).UTC()}, fakeRNG{}, nil)
	m2.Update(tea.KeyMsg{Type: tea.KeyDown})
	m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m2.machine.State() != app.FatalErrorState {
		t.Fatalf("expected fatal error state, got %s", m2.machine.State())
	}
}

func TestRootHandleBootBranches(t *testing.T) {
	now := time.Unix(1, 0).UTC()
	st := &fakeStore{err: errors.New("index error")}
	m := NewModel(st, fakeClock{now: now}, fakeRNG{}, nil)

	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.machine.State() != app.LoadGameListState {
		t.Fatalf("got %s", m.machine.State())
	}
	if !m.toast.Visible {
		t.Fatal("expected toast on index error")
	}

	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected quit cmd")
	}
}

func TestRootHandleNewGameCancelAndNoop(t *testing.T) {
	m := NewModel(&fakeStore{}, fakeClock{now: time.Unix(1, 0).UTC()}, fakeRNG{}, nil)
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.machine.State() != app.NewGameFormState {
		t.Fatalf("got %s", m.machine.State())
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.machine.State() != app.BootMenuState {
		t.Fatalf("got %s", m.machine.State())
	}
}

func TestRootHandleLoadGameCancelNewGameAndNoop(t *testing.T) {
	now := time.Unix(1, 0).UTC()
	st := &fakeStore{entries: []persistence.SaveIndexEntry{{SaveID: "1", CompanyName: "A", CompanyType: "SaaS", LastPlayedAt: now}}}
	m := NewModel(st, fakeClock{now: now}, fakeRNG{}, st.entries)
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.machine.State() != app.LoadGameListState {
		t.Fatalf("got %s", m.machine.State())
	}
	m.Update(struct{}{})
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.machine.State() != app.BootMenuState {
		t.Fatalf("got %s", m.machine.State())
	}

	m2 := NewModel(&fakeStore{}, fakeClock{now: now}, fakeRNG{}, nil)
	m2.Update(tea.KeyMsg{Type: tea.KeyDown})
	m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if m2.machine.State() != app.NewGameFormState {
		t.Fatalf("got %s", m2.machine.State())
	}
}

func TestRootHandleLoadingDefaultAndErrorBranches(t *testing.T) {
	now := time.Unix(1, 0).UTC()
	st := &fakeStore{touchErr: errors.New("touch failed")}
	m := NewModel(st, fakeClock{now: now}, fakeRNG{}, nil)
	_ = m.machine.Transition(app.EventGoNewGame)
	m.newGame.CompanyName = "Acme"
	m.newGame.FocusIndex = 2
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected submit cmd")
	}
	m.Update(struct{}{})
	m.Update(cmd())
	if !m.recoverModal.Open {
		t.Fatal("expected recovery modal")
	}
}

func TestRootHandleModalNonKeyAndEscToBoot(t *testing.T) {
	m := NewModel(&fakeStore{}, fakeClock{now: time.Unix(1, 0).UTC()}, fakeRNG{}, nil)
	_ = m.machine.Transition(app.EventGoNewGame)
	_ = m.machine.Transition(app.EventSubmitNewGame)
	m.loadingFrom = app.NewGameFormState
	m.openRecoverModal("err", []string{"Back to Menu"})
	if _, cmd := m.Update(struct{}{}); cmd != nil {
		t.Fatal("expected nil cmd for non-key modal message")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.machine.State() != app.BootMenuState {
		t.Fatalf("got %s", m.machine.State())
	}
}

func TestRootLoadSaveCmdTouchError(t *testing.T) {
	now := time.Unix(2, 0).UTC()
	st := &fakeStore{touchErr: errors.New("touch failed"), saves: map[string]persistence.SaveFile{"x": {SaveIdentity: persistence.SaveIdentity{SaveID: "x", CompanyName: "A", CompanyType: "SaaS", LastPlayedAt: now}}}}
	m := NewModel(st, fakeClock{now: now}, fakeRNG{}, nil)
	msg := m.loadSaveCmd("x")()
	if _, ok := msg.(loadResultMsg); !ok {
		t.Fatal("expected load result message")
	}
}

func TestRootViewModalAndToastAndAllStates(t *testing.T) {
	now := time.Unix(2, 0).UTC()
	m := NewModel(&fakeStore{}, fakeClock{now: now}, fakeRNG{}, nil)
	m.toast.Show("ok")
	m.openRecoverModal("err", []string{"Back to Menu"})
	_ = m.View()
	_ = m.machine.Transition(app.EventGoNewGame)
	_ = m.View()
	_ = m.machine.Transition(app.EventSubmitNewGame)
	_ = m.View()
	_ = m.machine.Transition(app.EventRecoverToBoot)
	_ = m.machine.Transition(app.EventGoLoadGame)
	_ = m.View()
	_ = m.machine.Transition(app.EventBack)
	_ = m.machine.Transition(app.EventGoHelp)
	_ = m.View()
	_ = m.machine.Transition(app.EventBack)
	_ = m.machine.Transition(app.EventShowFatalError)
	_ = m.View()
}
