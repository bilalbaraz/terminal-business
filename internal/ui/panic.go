package ui

import (
	"terminal-business/internal/ui/screens/fatal"

	tea "github.com/charmbracelet/bubbletea"
)

type PanicModel struct {
	fatal fatal.Model
}

func NewPanicModel(message string) *PanicModel {
	return &PanicModel{fatal: fatal.New(message)}
}

func (m *PanicModel) Init() tea.Cmd { return nil }

func (m *PanicModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.fatal.Update(msg) {
	case fatal.ActionQuit:
		return m, tea.Quit
	case fatal.ActionBack:
		return m, tea.Quit
	default:
		return m, nil
	}
}

func (m *PanicModel) View() string { return m.fatal.View() }
