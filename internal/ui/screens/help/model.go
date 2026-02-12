package help

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Action int

const (
	ActionNone Action = iota
	ActionBack
)

type Model struct {
	width  int
	height int
}

func New() Model { return Model{} }

func (m *Model) Update(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		if msg.String() == "esc" || msg.String() == "q" || msg.String() == "enter" {
			return ActionBack
		}
	}
	return ActionNone
}

func (m Model) View() string {
	body := []string{
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("45")).Render("Help"),
		"",
		"Controls:",
		"- arrows / j,k: navigate",
		"- enter: select",
		"- esc / q: back",
		"",
		"New Game creates a save and opens Dashboard.",
		"Load Game lists local saves as {Company Name} - {Company Type}.",
	}
	card := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Width(70).Render(strings.Join(body, "\n"))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card)
}
