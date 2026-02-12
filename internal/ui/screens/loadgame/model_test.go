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
	if a := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}); a != ActionNewGame {
		t.Fatalf("expected new game action, got %v", a)
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyEnter}); a != ActionNone {
		t.Fatalf("expected no action, got %v", a)
	}
}

func TestLoadGameSelectionAndCancel(t *testing.T) {
	now := time.Now()
	m := New([]persistence.SaveIndexEntry{{SaveID: "id1", CompanyName: "Acme", CompanyType: "SaaS", LastPlayedAt: now}})
	if a := m.Update(tea.KeyMsg{Type: tea.KeyEnter}); a != ActionSelect {
		t.Fatalf("expected select action, got %v", a)
	}
	if got := m.SelectedSaveID(); got != "id1" {
		t.Fatalf("got %s", got)
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyEsc}); a != ActionCancel {
		t.Fatalf("expected cancel action, got %v", a)
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}); a != ActionCancel {
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
