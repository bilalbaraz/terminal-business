package loadgame

import (
	"fmt"
	"strings"
	"time"

	"terminal-business/internal/persistence"
	"terminal-business/internal/ui/components"
	"terminal-business/internal/ui/layout"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ActionType int

const (
	ActionNone ActionType = iota
	ActionCancel
	ActionSelect
	ActionNewGame
	ActionDelete
)

type Action struct {
	Type   ActionType
	SaveID string
	Label  string
}

type Model struct {
	entries       []persistence.SaveIndexEntry
	menu          components.MenuList
	width         int
	height        int
	confirmDelete bool
	errorText     string
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
	if m.menu.Cursor >= len(m.menu.Items) {
		m.menu.Cursor = 0
	}
}

func (m *Model) SetError(text string) { m.errorText = text }

func (m Model) Empty() bool { return len(m.entries) == 0 }

func (m Model) SelectedSaveID() string {
	if len(m.entries) == 0 || m.menu.Cursor < 0 || m.menu.Cursor >= len(m.entries) {
		return ""
	}
	return m.entries[m.menu.Cursor].SaveID
}

func (m Model) SelectedLabel() string {
	if len(m.entries) == 0 || m.menu.Cursor < 0 || m.menu.Cursor >= len(m.entries) {
		return ""
	}
	e := m.entries[m.menu.Cursor]
	return fmt.Sprintf("%s - %s", e.CompanyName, e.CompanyType)
}

func (m *Model) Update(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		if m.confirmDelete {
			switch msg.String() {
			case "esc", "q":
				m.confirmDelete = false
				return Action{Type: ActionNone}
			case "enter":
				m.confirmDelete = false
				return Action{Type: ActionDelete, SaveID: m.SelectedSaveID(), Label: m.SelectedLabel()}
			}
			return Action{Type: ActionNone}
		}
		switch msg.String() {
		case "esc", "q":
			return Action{Type: ActionCancel}
		case "n", "N":
			if m.Empty() {
				return Action{Type: ActionNewGame}
			}
		case "d", "D":
			if !m.Empty() {
				m.confirmDelete = true
			}
		case "up", "k":
			m.menu.MoveUp()
		case "down", "j":
			m.menu.MoveDown()
		case "enter":
			if !m.Empty() {
				return Action{Type: ActionSelect}
			}
		}
	}
	return Action{Type: ActionNone}
}

func (m Model) View() string {
	regions := layout.Compute(m.width, m.height)
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("45")).Render("Load Game")
	if m.Empty() {
		body := title + "\n\nNo saved companies found\n\nPress n to create New Game or esc to return"
		if m.errorText != "" {
			body += "\n\n" + m.errorText
		}
		card := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Width(regions.MainWidth).Render(body)
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
	rows = layout.ClipRows(rows, regions.ScrollableRows)
	body := strings.Join(rows, "\n")
	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("enter: load • d: delete • ↑/↓ or j/k: navigate • esc/q: back")
	card := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Width(regions.Width - 4).Render(title + "\n\n" + body + "\n\n" + footer)
	if m.errorText != "" {
		card += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.errorText)
	}
	if m.confirmDelete {
		card += "\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render("Delete Save? "+m.SelectedLabel()+" (Enter=Delete Esc=Back)")
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card)
}
