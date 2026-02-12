package dashboard

import (
	"strings"
	"testing"

	domain "terminal-business/internal/domain/store"
	"terminal-business/internal/persistence"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDashboardNavigationBackAndView(t *testing.T) {
	m := New()
	m.SetCompany(persistence.SaveIdentity{CompanyName: "Acme", CompanyType: "SaaS"})
	m.SetGameState(domain.NewInitialState(1000, domain.DefaultCatalog(), domain.DefaultEconomyConfig()))
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m.Update(tea.KeyMsg{Type: tea.KeyUp})
	v := m.View()
	if !strings.Contains(v, "Acme") || !strings.Contains(v, "Sections") {
		t.Fatalf("unexpected view: %s", v)
	}
	if a := m.Update(tea.KeyMsg{Type: tea.KeyEsc}); a.Type != ActionBack {
		t.Fatalf("expected back action, got %+v", a)
	}
}

func TestDashboardStoreIntegrationBuyAction(t *testing.T) {
	m := New()
	m.SetStoreCatalog(domain.DefaultCatalog())
	m.SetStoreError("err")
	m.SetGameState(domain.NewInitialState(1000, domain.DefaultCatalog(), domain.DefaultEconomyConfig()))
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !strings.Contains(m.View(), "Store Items") {
		t.Fatal("expected store view")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	a := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if a.Type != ActionBuy {
		t.Fatalf("expected buy action, got %+v", a)
	}
	if a.ItemID == "" {
		t.Fatal("expected item id")
	}
	if b := m.Update(tea.KeyMsg{Type: tea.KeyEsc}); b.Type != ActionNone {
		t.Fatalf("expected none in store back, got %+v", b)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if strings.Contains(m.View(), "Store Items") {
		t.Fatal("should not enter store on non-store section")
	}
	if b := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}); b.Type != ActionBack {
		t.Fatalf("expected q back, got %+v", b)
	}
	if b := m.Update(struct{}{}); b.Type != ActionNone {
		t.Fatalf("expected none for unknown msg, got %+v", b)
	}
}
