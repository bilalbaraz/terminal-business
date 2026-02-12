package loadgame

import (
	"fmt"
	"strings"
	"time"

	"terminal-business/internal/persistence"
	"terminal-business/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Action int

const (
	ActionNone Action = iota
	ActionCancel
	ActionSelect
	ActionNewGame
)

type Model struct {
	entries []persistence.SaveIndexEntry
	menu    components.MenuList
	width   int
	height  int
}

func New(entries []persistence.SaveIndexEntry) Model {
	m := Model{}
	m.SetEntries(entries)
	return m
}

func (m *Model) SetEntries(entries []persistence.SaveIndexEntry) {
	m.entries = append([]persistence.SaveIndexEntry(nil), entries...)
	items := make([]string, 0, len(entries))
	for _, e := range entries {
		items = append(items, fmt.Sprintf("%s - %s", e.CompanyName, e.CompanyType))
	}
	m.menu = components.NewMenuList(items)
}

func (m Model) Empty() bool { return len(m.entries) == 0 }

func (m Model) SelectedSaveID() string {
	if len(m.entries) == 0 || m.menu.Cursor < 0 || m.menu.Cursor >= len(m.entries) {
		return ""
	}
	return m.entries[m.menu.Cursor].SaveID
}

func (m *Model) Update(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return ActionCancel
		case "n", "N":
			if m.Empty() {
				return ActionNewGame
			}
		case "up", "k":
			m.menu.MoveUp()
		case "down", "j":
			m.menu.MoveDown()
		case "enter":
			if !m.Empty() {
				return ActionSelect
			}
		}
	}
	return ActionNone
}

func (m Model) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("45")).Render("Load Game")
	if m.Empty() {
		card := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Width(60).Render(title + "\n\nNo saved companies found\n\nPress n to create New Game or esc to return")
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card)
	}
	rows := make([]string, len(m.menu.Items))
	for i, item := range m.menu.Items {
		ts := m.entries[i].LastPlayedAt.Format(time.DateTime)
		line := fmt.Sprintf("%s  %s", item, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(ts))
		if i == m.menu.Cursor {
			rows[i] = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render("› " + line)
		} else {
			rows[i] = "  " + line
		}
	}
	body := strings.Join(rows, "\n")
	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("enter: load • ↑/↓ or j/k: navigate • esc/q: back")
	card := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Width(80).Render(title + "\n\n" + body + "\n\n" + footer)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card)
}
