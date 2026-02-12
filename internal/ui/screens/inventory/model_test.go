package inventory

import (
	"strings"
	"testing"

	domain "terminal-business/internal/domain/store"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInventoryScreenBackAndView(t *testing.T) {
	catalog := domain.DefaultCatalog()
	m := New(catalog)
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m.SetState(domain.GameState{
		CompanyInventory: domain.NewInventoryFromEntries([]domain.InventoryItemInstance{
			{ItemID: domain.ItemDesk, Quantity: 1, RemainingDurabilityDays: 10},
			{ItemID: domain.ItemComputer, Quantity: 1, RemainingDurabilityDays: 0},
		}),
	})
	v := m.View()
	if !strings.Contains(v, "Company Inventory") || !strings.Contains(v, "Player Inventory") {
		t.Fatalf("unexpected view: %s", v)
	}
	if strings.Index(v, "Desk") > strings.Index(v, "Computer") {
		t.Fatalf("expected non-broken before broken, got: %s", v)
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyEsc}); a != ActionBack {
		t.Fatalf("expected back, got %v", a)
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}); a != ActionBack {
		t.Fatalf("expected q back, got %v", a)
	}
	if a := m.Update(struct{}{}); a != ActionNone {
		t.Fatalf("expected none, got %v", a)
	}
}

func TestInventoryScreenCatalogAndEmpty(t *testing.T) {
	m := New(domain.DefaultCatalog())
	m.SetCatalog(domain.NewCatalog(domain.CatalogConfig{}))
	m.SetState(domain.GameState{})
	v := m.View()
	if !strings.Contains(v, "(empty)") {
		t.Fatalf("unexpected view: %s", v)
	}
}
