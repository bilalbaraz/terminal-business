package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPanicModelLifecycle(t *testing.T) {
	m := NewPanicModel("boom")
	if m.Init() != nil {
		t.Fatal("expected nil init cmd")
	}
	if _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}); cmd == nil {
		t.Fatal("expected quit cmd")
	}
	if _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc}); cmd == nil {
		t.Fatal("expected quit cmd")
	}
	_ = m.View()
}
