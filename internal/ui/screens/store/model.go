package store

import (
	"fmt"
	"strings"

	domain "terminal-business/internal/domain/store"

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
	width       int
	height      int
	items       []domain.Item
	cursor      int
	confirmOpen bool
	errorText   string
}

func New(catalog domain.Catalog) Model {
	return Model{items: catalog.OrderedItems()}
}

func (m *Model) SetCatalog(catalog domain.Catalog) {
	m.items = catalog.OrderedItems()
	if len(m.items) == 0 {
		m.cursor = 0
		return
	}
	if m.cursor >= len(m.items) {
		m.cursor = len(m.items) - 1
	}
}

func (m *Model) SetError(text string) {
	m.errorText = text
}

func (m *Model) Update(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return Action{Type: ActionNone}
	case tea.KeyMsg:
		if m.confirmOpen {
			switch msg.String() {
			case "esc", "q":
				m.confirmOpen = false
				return Action{Type: ActionNone}
			case "enter":
				m.confirmOpen = false
				return Action{Type: ActionBuy, ItemID: m.SelectedItem().ItemID}
			}
			return Action{Type: ActionNone}
		}
		switch msg.String() {
		case "up", "k":
			if len(m.items) > 0 {
				m.cursor--
				if m.cursor < 0 {
					m.cursor = len(m.items) - 1
				}
			}
		case "down", "j":
			if len(m.items) > 0 {
				m.cursor = (m.cursor + 1) % len(m.items)
			}
		case "enter":
			if len(m.items) > 0 {
				m.confirmOpen = true
			}
		case "esc", "q":
			return Action{Type: ActionBack}
		}
	}
	return Action{Type: ActionNone}
}

func (m Model) SelectedItem() domain.Item {
	if len(m.items) == 0 {
		return domain.Item{}
	}
	if m.cursor < 0 || m.cursor >= len(m.items) {
		return domain.Item{}
	}
	return m.items[m.cursor]
}

func (m Model) View() string {
	if len(m.items) == 0 {
		return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Render("Store\n\nNo items available")
	}

	leftRows := make([]string, len(m.items))
	for i, it := range m.items {
		line := fmt.Sprintf("%s (%d)", it.DisplayName, it.Price)
		if i == m.cursor {
			leftRows[i] = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render("› " + line)
		} else {
			leftRows[i] = "  " + line
		}
	}
	left := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(28).Padding(1, 1).Render("Store Items\n\n" + strings.Join(leftRows, "\n"))

	selected := m.SelectedItem()
	right := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(46).Padding(1, 2).Render(
		fmt.Sprintf("%s\n\nCategory: %s\nPrice: %d\nDurability: %d days\nOwnership: %s\n\nEffects\n+Productivity: %d\n+Morale: %d\nTech Debt: %+d\nReputation: %+d",
			selected.DisplayName,
			selected.Category,
			selected.Price,
			selected.DurabilityDays,
			selected.Ownership,
			selected.Effects.ProductivityDelta,
			selected.Effects.MoraleDelta,
			selected.Effects.TechDebtDelta,
			selected.Effects.ReputationDelta,
		),
	)

	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("enter: buy • j/k or ↑/↓: select • esc/q: back")
	row := lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right)
	view := lipgloss.JoinVertical(lipgloss.Left, row, "", footer)
	if m.errorText != "" {
		view += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.errorText)
	}
	if m.confirmOpen {
		view += lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render("\n\nConfirm purchase? Enter=confirm Esc=cancel")
	}
	return view
}
