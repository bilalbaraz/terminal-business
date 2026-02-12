package fatal

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFatalActionsAndView(t *testing.T) {
	m := New("boom")
	if (&m).Init() != nil {
		t.Fatal("expected nil init")
	}
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if !strings.Contains(m.View(), "Fatal Error") {
		t.Fatal("expected fatal header")
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}); a != ActionBack {
		t.Fatalf("expected back action, got %v", a)
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyEsc}); a != ActionQuit {
		t.Fatalf("expected quit action, got %v", a)
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyDown}); a != ActionNone {
		t.Fatalf("expected none action, got %v", a)
	}
}
