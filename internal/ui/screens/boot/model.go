package boot

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
	ActionNewGame
	ActionLoadGame
	ActionHelp
	ActionExit
)

type Model struct {
	menu   components.MenuList
	width  int
	height int
}

func New() Model {
	return Model{menu: components.NewMenuList([]string{"New Game", "Load Game", "Help", "Exit"})}
}

func (m *Model) Update(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.menu.MoveUp()
		case "down", "j":
			m.menu.MoveDown()
		case "enter":
			switch m.menu.Selected() {
			case "New Game":
				return ActionNewGame
			case "Load Game":
				return ActionLoadGame
			case "Help":
				return ActionHelp
			case "Exit":
				return ActionExit
			}
		case "esc", "q":
			return ActionExit
		}
	}
	return ActionNone
}

func (m Model) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("45")).Render("TERMINAL BUSINESS")
	tagline := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("Build from zero to unicorn")

	lines := make([]string, 0, len(m.menu.Items))
	for i, item := range m.menu.Items {
		if i == m.menu.Cursor {
			lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render("› "+item))
		} else {
			lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("250")).Render("  "+item))
		}
	}
	body := strings.Join(lines, "\n")
	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("↑/↓ or j/k: navigate • enter: select • esc/q: exit")

	card := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 3).Width(44).Render(fmt.Sprintf("%s\n%s\n\n%s\n\n%s", title, tagline, body, footer))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card)
}
