package newgame

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewGameValidation(t *testing.T) {
	m := New()
	m.FocusIndex = 2
	if a := m.Update(tea.KeyMsg{Type: tea.KeyEnter}); a != ActionNone {
		t.Fatalf("expected no submit action, got %v", a)
	}
	if m.Error == "" {
		t.Fatal("expected validation error")
	}
}

func TestNewGameSubmitAndTypeControls(t *testing.T) {
	m := New()
	for _, r := range "Acme" {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m.FocusIndex = 1
	m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	if m.SelectedType() == "" {
		t.Fatal("expected selected type")
	}
	m.FocusIndex = 2
	if a := m.Update(tea.KeyMsg{Type: tea.KeyEnter}); a != ActionSubmit {
		t.Fatalf("expected submit action, got %v", a)
	}
}

func TestNewGameCancelBackspaceAndView(t *testing.T) {
	m := New()
	if a := m.Update(struct{}{}); a != ActionNone {
		t.Fatalf("expected none for unknown msg, got %v", a)
	}
	for _, r := range "AB" {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if m.CompanyName != "A" {
		t.Fatalf("got %q", m.CompanyName)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("tab")})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("shift+tab")})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("down")})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("up")})
	m.FocusIndex = 0
	m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.FocusIndex != 2 {
		t.Fatalf("expected reverse wrap to 2, got %d", m.FocusIndex)
	}
	m.FocusIndex = 1
	m.TypeIndex = 0
	m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if m.TypeIndex != len(m.CompanyTypes)-1 {
		t.Fatalf("expected left wrap, got %d", m.TypeIndex)
	}
	m.FocusIndex = 0
	m.TypeIndex = 1
	m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m.Update(tea.KeyMsg{Type: tea.KeyRight})
	if m.TypeIndex != 1 {
		t.Fatalf("type should not move when focus is not type field, got %d", m.TypeIndex)
	}
	m.FocusIndex = 0
	m.CompanyName = ""
	m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ab")})
	m.FocusIndex = 0
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.FocusIndex != 1 {
		t.Fatalf("expected focus advance, got %d", m.FocusIndex)
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyEsc}); a != ActionCancel {
		t.Fatalf("expected cancel action, got %v", a)
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}); a != ActionCancel {
		t.Fatalf("expected q cancel action, got %v", a)
	}
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	if !strings.Contains(m.View(), "Create New Company") {
		t.Fatal("view missing title")
	}
	m.FocusIndex = 2
	m.Error = "bad"
	if !strings.Contains(m.View(), "bad") {
		t.Fatal("view missing inline error")
	}
	m.FocusIndex = 0
	_ = m.View()
	m.FocusIndex = 1
	_ = m.View()
	m.TypeIndex = 99
	if m.SelectedType() != "" {
		t.Fatal("expected empty selected type")
	}
}
