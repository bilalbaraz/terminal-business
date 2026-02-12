package loadgame

import (
	"strings"
	"testing"
	"time"

	"terminal-business/internal/persistence"

	tea "github.com/charmbracelet/bubbletea"
)

func TestLoadGameEmptyStateShortcut(t *testing.T) {
	m := New(nil)
	if !m.Empty() {
		t.Fatal("expected empty")
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}); a.Type != ActionNewGame {
		t.Fatalf("expected new game action, got %v", a)
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyEnter}); a.Type != ActionNone {
		t.Fatalf("expected no action, got %v", a)
	}
}

func TestLoadGameSelectionAndCancel(t *testing.T) {
	now := time.Now()
	m := New([]persistence.SaveIndexEntry{{SaveID: "id1", CompanyName: "Acme", CompanyType: "SaaS", LastPlayedAt: now}})
	if a := m.Update(tea.KeyMsg{Type: tea.KeyEnter}); a.Type != ActionSelect {
		t.Fatalf("expected select action, got %v", a)
	}
	if got := m.SelectedSaveID(); got != "id1" {
		t.Fatalf("got %s", got)
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyEsc}); a.Type != ActionCancel {
		t.Fatalf("expected cancel action, got %v", a)
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}); a.Type != ActionCancel {
		t.Fatalf("expected q cancel action, got %v", a)
	}
}

func TestLoadGameNavigationWrapAndView(t *testing.T) {
	now := time.Now()
	m := New([]persistence.SaveIndexEntry{
		{SaveID: "id1", CompanyName: "A", CompanyType: "Game", LastPlayedAt: now},
		{SaveID: "id2", CompanyName: "B", CompanyType: "SaaS", LastPlayedAt: now},
	})
	m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if got := m.SelectedSaveID(); got != "id2" {
		t.Fatalf("got %s", got)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	v := m.View()
	if !strings.Contains(v, "Load Game") || !strings.Contains(v, "A - Game") {
		t.Fatalf("unexpected view: %s", v)
	}
	m.SetEntries(nil)
	if !strings.Contains(m.View(), "No saved companies found") {
		t.Fatal("expected empty state text")
	}
}

func TestLoadGameSelectedSaveIDOutOfRange(t *testing.T) {
	m := New([]persistence.SaveIndexEntry{{SaveID: "id1", CompanyName: "Acme", CompanyType: "SaaS", LastPlayedAt: time.Now()}})
	m.menu.Cursor = -1
	if got := m.SelectedSaveID(); got != "" {
		t.Fatalf("got %s", got)
	}
	m.menu.Cursor = 9
	if got := m.SelectedSaveID(); got != "" {
		t.Fatalf("got %s", got)
	}
}

func TestLoadGameDeleteConfirmFlow(t *testing.T) {
	now := time.Now()
	m := New([]persistence.SaveIndexEntry{{SaveID: "id1", CompanyName: "Acme", CompanyType: "SaaS", LastPlayedAt: now}})
	if a := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}); a.Type != ActionNone {
		t.Fatalf("expected none when opening delete confirm, got %v", a)
	}
	if !strings.Contains(m.View(), "Delete Save?") {
		t.Fatal("expected delete confirmation text")
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyEsc}); a.Type != ActionNone {
		t.Fatalf("expected none when canceling confirm, got %v", a)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	a := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if a.Type != ActionDelete || a.SaveID != "id1" {
		t.Fatalf("unexpected delete action %+v", a)
	}
}
