package help

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHelpBackAndView(t *testing.T) {
	m := New()
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if !strings.Contains(m.View(), "Help") {
		t.Fatal("expected help title")
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyEsc}); a != ActionBack {
		t.Fatalf("expected back action, got %v", a)
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyEnter}); a != ActionBack {
		t.Fatalf("expected enter back action, got %v", a)
	}
}
