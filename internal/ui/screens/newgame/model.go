package newgame

import (
	"fmt"
	"strings"

	"terminal-business/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Action int

const (
	ActionNone Action = iota
	ActionCancel
	ActionSubmit
)

type Model struct {
	CompanyName  string
	CompanyTypes []string
	TypeIndex    int
	FocusIndex   int
	Error        string
	width        int
	height       int
}

func New() Model {
	return Model{
		CompanyTypes: []string{"Game", "Fintech", "SaaS"},
		TypeIndex:    2,
	}
}

func (m *Model) Update(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return ActionCancel
		case "tab", "down", "j":
			m.FocusIndex = (m.FocusIndex + 1) % 3
		case "shift+tab", "up", "k":
			m.FocusIndex--
			if m.FocusIndex < 0 {
				m.FocusIndex = 2
			}
		case "left", "h":
			if m.FocusIndex == 1 {
				m.TypeIndex--
				if m.TypeIndex < 0 {
					m.TypeIndex = len(m.CompanyTypes) - 1
				}
			}
		case "right", "l":
			if m.FocusIndex == 1 {
				m.TypeIndex = (m.TypeIndex + 1) % len(m.CompanyTypes)
			}
		case "enter":
			if m.FocusIndex == 2 {
				m.Error = components.ValidateCompanyName(m.CompanyName)
				if m.Error != "" {
					return ActionNone
				}
				return ActionSubmit
			}
			m.FocusIndex = (m.FocusIndex + 1) % 3
		case "backspace":
			if m.FocusIndex == 0 && len(m.CompanyName) > 0 {
				runes := []rune(m.CompanyName)
				m.CompanyName = string(runes[:len(runes)-1])
			}
		default:
			if m.FocusIndex == 0 && len(msg.Runes) == 1 {
				m.CompanyName += string(msg.Runes)
			}
		}
	}
	return ActionNone
}

func (m Model) SelectedType() string {
	if m.TypeIndex < 0 || m.TypeIndex >= len(m.CompanyTypes) {
		return ""
	}
	return m.CompanyTypes[m.TypeIndex]
}

func (m Model) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("45")).Render("Create New Company")

	nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	if m.FocusIndex == 0 {
		nameStyle = nameStyle.Bold(true).Foreground(lipgloss.Color("205"))
	}
	typeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	if m.FocusIndex == 1 {
		typeStyle = typeStyle.Bold(true).Foreground(lipgloss.Color("205"))
	}
	submitStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	if m.FocusIndex == 2 {
		submitStyle = submitStyle.Bold(true).Foreground(lipgloss.Color("205"))
	}
	parts := []string{title, "", nameStyle.Render("Company Name"), "  " + m.CompanyName}
	if m.Error != "" {
		parts = append(parts, lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("  "+m.Error))
	}
	parts = append(parts, "", typeStyle.Render("Company Type"), "  "+m.renderTypes(), "", submitStyle.Render("[ Create Company ]"), "", lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("tab/j/k: move • h/l: change type • enter: continue • esc/q: cancel"))
	card := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Width(60).Render(strings.Join(parts, "\n"))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card)
}

func (m Model) renderTypes() string {
	items := make([]string, len(m.CompanyTypes))
	for i, t := range m.CompanyTypes {
		prefix := "( )"
		if i == m.TypeIndex {
			prefix = "(*)"
		}
		items[i] = fmt.Sprintf("%s %s", prefix, t)
	}
	return strings.Join(items, "   ")
}
