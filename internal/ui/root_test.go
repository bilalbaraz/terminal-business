package ui

import (
	"context"
	"errors"
	"testing"
	"time"

	"terminal-business/internal/app"
	"terminal-business/internal/persistence"

	tea "github.com/charmbracelet/bubbletea"
)

type fakeClock struct{ now time.Time }

func (f fakeClock) Now() time.Time { return f.now }

type fakeRNG struct{}

func (fakeRNG) Int63() int64 { return 42 }

type fakeStore struct {
	entries   []persistence.SaveIndexEntry
	saves     map[string]persistence.SaveFile
	err       error
	loadErr   error
	touchErr  error
	createErr error
}

func (f *fakeStore) LoadIndex(context.Context) ([]persistence.SaveIndexEntry, error) {
	return f.entries, f.err
}

func (f *fakeStore) CreateSave(_ context.Context, s persistence.SaveFile) (persistence.SaveFile, error) {
	if f.createErr != nil {
		return persistence.SaveFile{}, f.createErr
	}
	if f.err != nil {
		return persistence.SaveFile{}, f.err
	}
	s.SaveIdentity.SaveID = "sid"
	s.SaveIdentity.CompanyID = "cid"
	if f.saves == nil {
		f.saves = map[string]persistence.SaveFile{}
	}
	f.saves[s.SaveIdentity.SaveID] = s
	return s, nil
}

func (f *fakeStore) LoadSave(_ context.Context, id string) (persistence.SaveFile, error) {
	if f.loadErr != nil {
		return persistence.SaveFile{}, f.loadErr
	}
	s, ok := f.saves[id]
	if !ok {
		return persistence.SaveFile{}, errors.New("not found")
	}
	return s, nil
}

func (f *fakeStore) TouchSave(_ context.Context, id string, t time.Time) error {
	if f.touchErr != nil {
		return f.touchErr
	}
	if f.err != nil {
		return f.err
	}
	s := f.saves[id]
	s.SaveIdentity.LastPlayedAt = t
	f.saves[id] = s
	return nil
}

func TestRootInitAndViewAcrossStates(t *testing.T) {
	st := &fakeStore{}
	m := NewModel(st, fakeClock{now: time.Unix(100, 0).UTC()}, fakeRNG{}, nil)
	if m.Init() != nil {
		t.Fatal("expected nil init")
	}
	_ = m.View()
	_ = m.machine.Transition(app.EventGoNewGame)
	_ = m.View()
	_ = m.machine.Transition(app.EventSubmitNewGame)
	_ = m.View()
	_ = m.machine.Transition(app.EventLoadSucceeded)
	_ = m.View()
	_ = m.machine.Transition(app.EventShowFatalError)
	_ = m.View()
}

func TestRootFlowNewGameToDashboard(t *testing.T) {
	st := &fakeStore{}
	m := NewModel(st, fakeClock{now: time.Unix(100, 0).UTC()}, fakeRNG{}, nil)
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(*Model)
	if m.machine.State() != app.NewGameFormState {
		t.Fatalf("got state %s", m.machine.State())
	}
	for _, r := range "Acme" {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m.newGame.FocusIndex = 2
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(*Model)
	if m.machine.State() != app.LoadingGameState {
		t.Fatalf("got state %s", m.machine.State())
	}
	msg := cmd()
	model, _ = m.Update(msg)
	m = model.(*Model)
	if m.machine.State() != app.DashboardState {
		t.Fatalf("got state %s", m.machine.State())
	}
}

func TestRootFlowCreateFailureToModalRecovery(t *testing.T) {
	now := time.Unix(100, 0).UTC()
	st := &fakeStore{createErr: errors.New("create failed")}
	m := NewModel(st, fakeClock{now: now}, fakeRNG{}, nil)
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	for _, r := range "Acme" {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m.newGame.FocusIndex = 2
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m.Update(cmd())
	if !m.recoverModal.Open {
		t.Fatal("expected modal open")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.machine.State() != app.BootMenuState {
		t.Fatalf("got state %s", m.machine.State())
	}
}

func TestRootFlowLoadFailureToModalRecovery(t *testing.T) {
	now := time.Unix(100, 0).UTC()
	st := &fakeStore{
		entries: []persistence.SaveIndexEntry{{SaveID: "s1", CompanyName: "A", CompanyType: "SaaS", LastPlayedAt: now}},
		saves: map[string]persistence.SaveFile{
			"s1": {SaveIdentity: persistence.SaveIdentity{SaveID: "s1", CompanyName: "A", CompanyType: "SaaS", LastPlayedAt: now}},
		},
		loadErr: errors.New("boom"),
	}
	m := NewModel(st, fakeClock{now: now}, fakeRNG{}, st.entries)
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(*Model)
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(*Model)
	if m.machine.State() != app.LoadGameListState {
		t.Fatalf("got state %s", m.machine.State())
	}
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(*Model)
	if m.machine.State() != app.LoadingGameState {
		t.Fatalf("got state %s", m.machine.State())
	}
	model, _ = m.Update(cmd())
	m = model.(*Model)
	if !m.recoverModal.Open {
		t.Fatal("expected modal open")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m.Update(tea.KeyMsg{Type: tea.KeyRight})
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(*Model)
	if m.machine.State() != app.LoadGameListState && m.machine.State() != app.BootMenuState {
		t.Fatalf("unexpected state %s", m.machine.State())
	}
}

func TestRootHelpDashboardAndFatalHandlers(t *testing.T) {
	now := time.Unix(100, 0).UTC()
	st := &fakeStore{entries: []persistence.SaveIndexEntry{{SaveID: "s1", CompanyName: "A", CompanyType: "SaaS", LastPlayedAt: now}}, saves: map[string]persistence.SaveFile{"s1": {SaveIdentity: persistence.SaveIdentity{SaveID: "s1", CompanyName: "A", CompanyType: "SaaS", LastPlayedAt: now}}}}
	m := NewModel(st, fakeClock{now: now}, fakeRNG{}, st.entries)

	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.machine.State() != app.HelpState {
		t.Fatalf("got state %s", m.machine.State())
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.machine.State() != app.BootMenuState {
		t.Fatalf("got state %s", m.machine.State())
	}

	m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m.Update(cmd())
	if m.machine.State() != app.DashboardState {
		t.Fatalf("got state %s", m.machine.State())
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.machine.State() != app.BootMenuState {
		t.Fatalf("got state %s", m.machine.State())
	}

	_ = m.machine.Transition(app.EventShowFatalError)
	m.fatal = m.fatal
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	if m.machine.State() != app.BootMenuState {
		t.Fatalf("got state %s", m.machine.State())
	}
	_ = m.machine.Transition(app.EventShowFatalError)
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected quit cmd")
	}
}

func TestRootHandleModalEscPath(t *testing.T) {
	m := NewModel(&fakeStore{}, fakeClock{now: time.Unix(100, 0).UTC()}, fakeRNG{}, nil)
	_ = m.machine.Transition(app.EventGoNewGame)
	_ = m.machine.Transition(app.EventSubmitNewGame)
	m.loadingFrom = app.LoadGameListState
	m.openRecoverModal("error", []string{"Back to List", "Back to Menu"})
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.machine.State() != app.LoadGameListState {
		t.Fatalf("got state %s", m.machine.State())
	}
}
