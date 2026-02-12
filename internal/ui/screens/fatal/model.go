package fatal

import (
	"terminal-business/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Action int

const (
	ActionNone Action = iota
	ActionBack
	ActionQuit
)

type Model struct {
	Message string
	width   int
	height  int
}

func New(message string) Model { return Model{Message: message} }

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "b", "B":
			return ActionBack
		case "q", "esc", "ctrl+c":
			return ActionQuit
		}
	}
	return ActionNone
}

func (m Model) View() string {
	card := lipgloss.NewStyle().Border(lipgloss.ThickBorder()).BorderForeground(lipgloss.Color("196")).Padding(1, 2).Width(70).Render(components.ErrorScreen("Fatal Error", m.Message))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card)
}
