package boot

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestBootMenuNavigationAndSelection(t *testing.T) {
	m := New()
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	if a := m.Update(tea.KeyMsg{Type: tea.KeyDown}); a != ActionNone {
		t.Fatalf("unexpected action %v", a)
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyEnter}); a != ActionLoadGame {
		t.Fatalf("expected load game action, got %v", a)
	}
}

func TestBootMenuSelectAllItemsAndExitShortcut(t *testing.T) {
	m := New()
	if a := m.Update(tea.KeyMsg{Type: tea.KeyEnter}); a != ActionNewGame {
		t.Fatalf("expected new game action, got %v", a)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if a := m.Update(tea.KeyMsg{Type: tea.KeyEnter}); a != ActionHelp {
		t.Fatalf("expected help action, got %v", a)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if a := m.Update(tea.KeyMsg{Type: tea.KeyEnter}); a != ActionExit {
		t.Fatalf("expected exit action, got %v", a)
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyEsc}); a != ActionExit {
		t.Fatalf("expected esc exit action, got %v", a)
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}); a != ActionExit {
		t.Fatalf("expected q exit action, got %v", a)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
}

func TestBootView(t *testing.T) {
	m := New()
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 35})
	v := m.View()
	if !strings.Contains(v, "TERMINAL BUSINESS") {
		t.Fatalf("unexpected view: %s", v)
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}); a != ActionNone {
		t.Fatalf("unexpected action %v", a)
	}
}
