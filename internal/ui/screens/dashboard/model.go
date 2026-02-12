package dashboard

import (
	"fmt"
	"strings"

	domain "terminal-business/internal/domain/store"
	"terminal-business/internal/persistence"
	"terminal-business/internal/ui/layout"
	uiinventory "terminal-business/internal/ui/screens/inventory"
	uimarket "terminal-business/internal/ui/screens/market"
	uistore "terminal-business/internal/ui/screens/store"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ActionType int

const (
	ActionNone ActionType = iota
	ActionBack
	ActionBuy
	ActionAcceptJob
	ActionAdvanceDay
)

type Action struct {
	Type   ActionType
	ItemID domain.ItemID
	JobID  string
}

type Model struct {
	Company   persistence.SaveIdentity
	GameState domain.GameState

	width       int
	height      int
	cursor      int
	nav         []string
	inStore     bool
	inInventory bool
	inMarket    bool
	store       uistore.Model
	inventory   uiinventory.Model
	market      uimarket.Model
}

func New() Model {
	catalog := domain.DefaultCatalog()
	return Model{
		nav:       []string{"Store", "Market", "Inventory", "HR", "Investors", "Reports", "Settings"},
		store:     uistore.New(catalog),
		inventory: uiinventory.New(catalog),
		market:    uimarket.New(),
	}
}

func (m *Model) SetCompany(company persistence.SaveIdentity) {
	m.Company = company
}

func (m *Model) SetGameState(state domain.GameState) {
	m.GameState = state
	m.inventory.SetState(state)
	m.market.SetActiveAndCompleted(state.ActiveJobs, state.CompletedJobs)
}

func (m *Model) SetStoreCatalog(catalog domain.Catalog) {
	m.store.SetCatalog(catalog)
	m.inventory.SetCatalog(catalog)
}

func (m *Model) SetStoreError(msg string) {
	m.store.SetError(msg)
}

func (m *Model) SetMarketJobs(jobs []domain.JobOffer) {
	m.market.SetJobs(jobs)
}

func (m *Model) SetMarketError(msg string) { m.market.SetError(msg) }

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

	if m.inInventory {
		if m.inventory.Update(msg) == uiinventory.ActionBack {
			m.inInventory = false
		}
		return Action{Type: ActionNone}
	}

	if m.inMarket {
		a := m.market.Update(msg)
		switch a.Type {
		case uimarket.ActionBack:
			m.inMarket = false
			return Action{Type: ActionNone}
		case uimarket.ActionAccept:
			return Action{Type: ActionAcceptJob, JobID: a.JobID}
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
			if m.nav[m.cursor] == "Inventory" {
				m.inInventory = true
			}
			if m.nav[m.cursor] == "Market" {
				m.inMarket = true
			}
		case "n":
			return Action{Type: ActionAdvanceDay}
		case "esc", "q":
			return Action{Type: ActionBack}
		}
	}
	return Action{Type: ActionNone}
}

func (m Model) View() string {
	regions := layout.Compute(m.width, m.height)
	head := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("45")).Render(fmt.Sprintf("%s (%s)", m.Company.CompanyName, m.Company.CompanyType))
	if m.inStore {
		stats := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(regions.RightWidth).Padding(1, 1).Render(fmt.Sprintf("Stats\n\nDay: %d\nCash: %d\nBurn: %d\nRunway: %.2f", m.GameState.Day, m.GameState.Metrics.Cash, m.GameState.Metrics.BurnRate, m.GameState.Metrics.RunwayMonths))
		row := lipgloss.JoinHorizontal(lipgloss.Top, m.store.View(), " ", stats)
		footer := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Store: enter buy • esc back")
		content := lipgloss.JoinVertical(lipgloss.Left, head, "", row, "", footer)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}
	if m.inInventory {
		footer := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Inventory: esc back")
		content := lipgloss.JoinVertical(lipgloss.Left, head, "", m.inventory.View(), "", footer)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}
	if m.inMarket {
		footer := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Market: enter/a accept • esc back")
		content := lipgloss.JoinVertical(lipgloss.Left, head, "", m.market.View(m.GameState.Day), "", footer)
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
	operational := "Operational: YES"
	if !domain.IsOperational(m.GameState) {
		operational = "Operational: NO (buy Desk + Chair + Computer)"
	}
	activeJob := "Active Job: none"
	if len(m.GameState.ActiveJobs) > 0 {
		j := m.GameState.ActiveJobs[0]
		activeJob = fmt.Sprintf("Active Job: %s (%dd left)", j.Title, domain.ActiveJobCountdown(m.GameState.Day, j))
	}

	if regions.Mode == layout.CompactMode {
		tabs := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(regions.MainWidth).Padding(1, 1).Render("Tabs\n\n" + strings.Join(m.nav, " | "))
		main := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(regions.MainWidth).Padding(1, 2).Render(operational + "\n" + activeJob)
		stats := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(regions.MainWidth).Padding(1, 1).Render(fmt.Sprintf("Stats\n\nDay: %d\nCash: %d\nBurn: %d\nRunway: %.2f", m.GameState.Day, m.GameState.Metrics.Cash, m.GameState.Metrics.BurnRate, m.GameState.Metrics.RunwayMonths))
		footer := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("j/k navigate • enter open • n next day • esc/q back")
		content := lipgloss.JoinVertical(lipgloss.Left, head, "", tabs, "", main, "", stats, "", footer)
		return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, content)
	}

	sidebar := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(regions.SidebarWidth).Padding(1, 1).Render("Sections\n\n" + strings.Join(rows, "\n"))
	main := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(regions.MainWidth).Padding(1, 2).Render("Main Panel\n\nSelect Store, Market, or Inventory.\n\n" + operational + "\n" + activeJob)
	stats := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(regions.RightWidth).Padding(1, 1).Render(fmt.Sprintf("Stats\n\nDay: %d\nCash: %d\nBurn: %d\nRunway: %.2f", m.GameState.Day, m.GameState.Metrics.Cash, m.GameState.Metrics.BurnRate, m.GameState.Metrics.RunwayMonths))
	row := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, " ", main, " ", stats)
	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("j/k navigate • enter open • n next day • esc/q back")
	content := lipgloss.JoinVertical(lipgloss.Left, head, "", row, "", footer)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
