package market

import (
	"fmt"
	"strings"

	domain "terminal-business/internal/domain/store"
	"terminal-business/internal/ui/layout"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ActionType int

const (
	ActionNone ActionType = iota
	ActionBack
	ActionAccept
)

type Action struct {
	Type  ActionType
	JobID string
}

type Model struct {
	width       int
	height      int
	jobs        []domain.JobOffer
	active      []domain.ActiveJob
	completed   []domain.CompletedJob
	cursor      int
	confirmOpen bool
	errorText   string
	roiIndex    int
}

func New() Model { return Model{roiIndex: -1} }

func (m *Model) SetJobs(jobs []domain.JobOffer) {
	m.jobs = append([]domain.JobOffer(nil), jobs...)
	if len(m.jobs) == 0 {
		m.cursor = 0
	} else if m.cursor >= len(m.jobs) {
		m.cursor = len(m.jobs) - 1
	}
	m.roiIndex = domain.BestROISoonestIndex(m.jobs)
}

func (m *Model) SetActiveAndCompleted(active []domain.ActiveJob, completed []domain.CompletedJob) {
	m.active = append([]domain.ActiveJob(nil), active...)
	m.completed = append([]domain.CompletedJob(nil), completed...)
}

func (m *Model) SetError(text string) { m.errorText = text }

func (m *Model) SelectedJobID() string {
	if len(m.jobs) == 0 || m.cursor < 0 || m.cursor >= len(m.jobs) {
		return ""
	}
	return m.jobs[m.cursor].JobID
}

func (m *Model) Update(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		if m.confirmOpen {
			switch msg.String() {
			case "esc", "q":
				m.confirmOpen = false
			case "enter":
				m.confirmOpen = false
				return Action{Type: ActionAccept, JobID: m.SelectedJobID()}
			}
			return Action{Type: ActionNone}
		}
		switch msg.String() {
		case "up", "k":
			if len(m.jobs) > 0 {
				m.cursor--
				if m.cursor < 0 {
					m.cursor = len(m.jobs) - 1
				}
			}
		case "down", "j":
			if len(m.jobs) > 0 {
				m.cursor = (m.cursor + 1) % len(m.jobs)
			}
		case "enter", "a":
			if len(m.jobs) > 0 {
				m.confirmOpen = true
			}
		case "esc", "q":
			return Action{Type: ActionBack}
		}
	}
	return Action{Type: ActionNone}
}

func (m Model) View(currentTick int) string {
	regions := layout.Compute(m.width, m.height)
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("45")).Render("Freelance Market")
	if len(m.jobs) == 0 {
		body := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(regions.MainWidth).Padding(1, 2).Render(title + "\n\nNo jobs available")
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, body)
	}

	rows := make([]string, len(m.jobs))
	for i, job := range m.jobs {
		marker := " "
		if i == m.roiIndex {
			marker = "*"
		}
		line := fmt.Sprintf("%s %s | payout:%d | %dd", marker, job.Title, job.Payout, job.DurationDays)
		if i == m.cursor {
			rows[i] = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render("› " + line)
		} else {
			rows[i] = "  " + line
		}
	}
	rows = layout.ClipRows(rows, regions.ScrollableRows)
	list := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(regions.MainWidth).Padding(1, 1).Render("Jobs\n\n" + strings.Join(rows, "\n"))

	selected := m.jobs[m.cursor]
	details := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(regions.RightWidth).Padding(1, 1).Render(
		fmt.Sprintf("Details\n\n%s\nPayout: %d\nDuration: %d days\nStatus: %s", selected.Title, selected.Payout, selected.DurationDays, selected.Status),
	)

	activeLine := "None"
	if len(m.active) > 0 {
		first := m.active[0]
		activeLine = fmt.Sprintf("%s (%dd left)", first.Title, domain.ActiveJobCountdown(currentTick, first))
	}
	historyRows := []string{"Completed"}
	for _, h := range layout.ClipRows(m.completed, 2) {
		historyRows = append(historyRows, fmt.Sprintf("%s +%d", h.Title, h.Payout))
	}
	info := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(regions.RightWidth).Padding(1, 1).Render("Active\n\n" + activeLine + "\n\n" + strings.Join(historyRows, "\n"))

	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("enter/a: accept • esc/q: back • * best ROI")
	if m.errorText != "" {
		footer += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.errorText)
	}
	if m.confirmOpen {
		footer += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render("Accept selected job? Enter=confirm Esc=cancel")
	}

	if regions.Mode == layout.CompactMode {
		stack := lipgloss.JoinVertical(lipgloss.Left, title, "", list, "", details, "", info, "", footer)
		return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, stack)
	}

	right := lipgloss.JoinVertical(lipgloss.Left, details, "", info)
	row := lipgloss.JoinHorizontal(lipgloss.Top, list, " ", right)
	content := lipgloss.JoinVertical(lipgloss.Left, title, "", row, "", footer)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
