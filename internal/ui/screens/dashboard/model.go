package dashboard

import (
	"fmt"
	"strings"

	domain "terminal-business/internal/domain/store"
	"terminal-business/internal/persistence"
	uistore "terminal-business/internal/ui/screens/store"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ActionType int

const (
	ActionNone ActionType = iota
	ActionBack
	ActionBuy
)

type Action struct {
	Type   ActionType
	ItemID domain.ItemID
}

type Model struct {
	Company   persistence.SaveIdentity
	GameState domain.GameState

	width   int
	height  int
	cursor  int
	nav     []string
	inStore bool
	store   uistore.Model
}

func New() Model {
	catalog := domain.DefaultCatalog()
	return Model{
		nav:   []string{"Store", "HR", "Investors", "Reports", "Settings"},
		store: uistore.New(catalog),
	}
}

func (m *Model) SetCompany(company persistence.SaveIdentity) {
	m.Company = company
}

func (m *Model) SetGameState(state domain.GameState) {
	m.GameState = state
}

func (m *Model) SetStoreCatalog(catalog domain.Catalog) {
	m.store.SetCatalog(catalog)
}

func (m *Model) SetStoreError(msg string) {
	m.store.SetError(msg)
}

func (m *Model) Update(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	}

	if m.inStore {
		a := m.store.Update(msg)
		switch a.Type {
		case uistore.ActionBack:
			m.inStore = false
			return Action{Type: ActionNone}
		case uistore.ActionBuy:
			return Action{Type: ActionBuy, ItemID: a.ItemID}
		default:
			return Action{Type: ActionNone}
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.nav) - 1
			}
		case "down", "j":
			m.cursor = (m.cursor + 1) % len(m.nav)
		case "enter":
			if m.nav[m.cursor] == "Store" {
				m.inStore = true
			}
		case "esc", "q":
			return Action{Type: ActionBack}
		}
	}
	return Action{Type: ActionNone}
}

func (m Model) View() string {
	head := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("45")).Render(fmt.Sprintf("%s (%s)", m.Company.CompanyName, m.Company.CompanyType))
	if m.inStore {
		stats := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(24).Padding(1, 1).Render(fmt.Sprintf("Stats\n\nCash: %d\nBurn: %d\nRunway: %.2f", m.GameState.Metrics.Cash, m.GameState.Metrics.BurnRate, m.GameState.Metrics.RunwayMonths))
		row := lipgloss.JoinHorizontal(lipgloss.Top, m.store.View(), " ", stats)
		footer := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Store: enter buy • esc back")
		content := lipgloss.JoinVertical(lipgloss.Left, head, "", row, "", footer)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	rows := make([]string, len(m.nav))
	for i, item := range m.nav {
		if i == m.cursor {
			rows[i] = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render("› " + item)
		} else {
			rows[i] = "  " + item
		}
	}
	sidebar := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(22).Padding(1, 1).Render("Sections\n\n" + strings.Join(rows, "\n"))
	main := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(48).Padding(1, 2).Render("Main Panel\n\nSelect Store to purchase equipment.")
	stats := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(24).Padding(1, 1).Render(fmt.Sprintf("Stats\n\nCash: %d\nBurn: %d\nRunway: %.2f", m.GameState.Metrics.Cash, m.GameState.Metrics.BurnRate, m.GameState.Metrics.RunwayMonths))

	row := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, " ", main, " ", stats)
	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("j/k: navigate sidebar • enter: open • esc/q: back to menu")
	content := lipgloss.JoinVertical(lipgloss.Left, head, "", row, "", footer)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
