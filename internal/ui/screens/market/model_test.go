package market

import (
	"strings"
	"testing"

	domain "terminal-business/internal/domain/store"

	tea "github.com/charmbracelet/bubbletea"
)

func sampleJobs() []domain.JobOffer {
	return []domain.JobOffer{
		{JobID: "j1", Title: "A", Payout: 100, DurationDays: 2, Status: domain.JobAvailable},
		{JobID: "j2", Title: "B", Payout: 200, DurationDays: 2, Status: domain.JobAvailable},
	}
}

func TestMarketSelectionAcceptAndBack(t *testing.T) {
	m := New()
	m.SetJobs(sampleJobs())
	m.SetActiveAndCompleted([]domain.ActiveJob{{Title: "X", DueAtTick: 4}}, []domain.CompletedJob{{Title: "Done", Payout: 50}})
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if a := m.Update(tea.KeyMsg{Type: tea.KeyDown}); a.Type != ActionNone {
		t.Fatalf("unexpected action %+v", a)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	a := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if a.Type != ActionAccept || a.JobID == "" {
		t.Fatalf("unexpected accept %+v", a)
	}
	if b := m.Update(tea.KeyMsg{Type: tea.KeyEsc}); b.Type != ActionBack {
		t.Fatalf("expected back got %+v", b)
	}
}

func TestMarketConfirmCancelAndViewModes(t *testing.T) {
	m := New()
	m.SetJobs(sampleJobs())
	m.SetError("no capacity")
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if a := m.Update(tea.KeyMsg{Type: tea.KeyUp}); a.Type != ActionNone {
		t.Fatalf("expected none in confirm got %+v", a)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m.Update(tea.WindowSizeMsg{Width: 70, Height: 15})
	v := m.View(1)
	if !strings.Contains(v, "Freelance Market") || !strings.Contains(v, "no capacity") {
		t.Fatalf("unexpected view %s", v)
	}
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	_ = m.View(1)
	if m.SelectedJobID() == "" {
		t.Fatal("expected selected job")
	}
	m.SetJobs(nil)
	if m.SelectedJobID() != "" {
		t.Fatal("expected empty selected id")
	}
	if !strings.Contains(m.View(1), "No jobs available") {
		t.Fatal("expected empty view")
	}
}
