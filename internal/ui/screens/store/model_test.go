package store

import (
	"strings"
	"testing"

	domain "terminal-business/internal/domain/store"

	tea "github.com/charmbracelet/bubbletea"
)

func TestStoreSelectionAndBuyFlow(t *testing.T) {
	m := New(domain.DefaultCatalog())
	if got := m.SelectedItem().ItemID; got == "" {
		t.Fatal("expected selected item")
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyDown}); a.Type != ActionNone {
		t.Fatalf("unexpected action: %+v", a)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	a := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if a.Type != ActionBuy {
		t.Fatalf("expected buy action, got %+v", a)
	}
	if a.ItemID == "" {
		t.Fatal("expected item id")
	}
}

func TestStoreConfirmCancelAndBack(t *testing.T) {
	m := New(domain.DefaultCatalog())
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if a := m.Update(tea.KeyMsg{Type: tea.KeyEsc}); a.Type != ActionNone {
		t.Fatalf("expected none, got %+v", a)
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyEsc}); a.Type != ActionBack {
		t.Fatalf("expected back, got %+v", a)
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}); a.Type != ActionBack {
		t.Fatalf("expected q back, got %+v", a)
	}
}

func TestStoreCatalogSetErrorAndView(t *testing.T) {
	m := New(domain.DefaultCatalog())
	m.SetError("boom")
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if !strings.Contains(m.View(), "boom") {
		t.Fatal("expected error in view")
	}
	m.SetCatalog(domain.NewCatalog(domain.CatalogConfig{}))
	if !strings.Contains(m.View(), "No items available") {
		t.Fatal("expected empty catalog view")
	}
	if got := m.SelectedItem().ItemID; got != "" {
		t.Fatalf("expected empty selected item, got %s", got)
	}
	m.SetCatalog(domain.DefaultCatalog())
	m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if got := m.SelectedItem().ItemID; got == "" {
		t.Fatal("expected wrapped selection")
	}
	m.cursor = 99
	m.SetCatalog(domain.DefaultCatalog())
	if m.cursor != 2 {
		t.Fatalf("expected cursor clamp to 2, got %d", m.cursor)
	}
	m.cursor = -1
	if got := m.SelectedItem().ItemID; got != "" {
		t.Fatalf("expected empty selected item, got %s", got)
	}
	m.cursor = 99
	if got := m.SelectedItem().ItemID; got != "" {
		t.Fatalf("expected empty selected item, got %s", got)
	}
	m = New(domain.DefaultCatalog())
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if a := m.Update(tea.KeyMsg{Type: tea.KeyUp}); a.Type != ActionNone {
		t.Fatalf("expected none for ignored key in confirm, got %+v", a)
	}
	if !strings.Contains(m.View(), "Confirm purchase?") {
		t.Fatal("expected confirm text in view")
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}); a.Type != ActionNone {
		t.Fatalf("expected none while closing confirm, got %+v", a)
	}
	m.items = nil
	if a := m.Update(tea.KeyMsg{Type: tea.KeyEnter}); a.Type != ActionNone {
		t.Fatalf("expected none with empty items, got %+v", a)
	}
	if a := m.Update(struct{}{}); a.Type != ActionNone {
		t.Fatalf("expected none for unknown msg, got %+v", a)
	}
}
