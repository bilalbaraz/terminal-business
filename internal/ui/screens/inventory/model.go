package inventory

import (
	"fmt"
	"sort"
	"strings"

	domain "terminal-business/internal/domain/store"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Action int

const (
	ActionNone Action = iota
	ActionBack
)

type row struct {
	name       string
	quantity   int
	durability int
	broken     bool
}

type Model struct {
	width   int
	height  int
	catalog domain.Catalog
	state   domain.GameState
}

func New(catalog domain.Catalog) Model {
	return Model{catalog: catalog}
}

func (m *Model) SetCatalog(catalog domain.Catalog) {
	m.catalog = catalog
}

func (m *Model) SetState(state domain.GameState) {
	m.state = state
}

func (m *Model) Update(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return ActionBack
		}
	}
	return ActionNone
}

func (m Model) View() string {
	company := renderSection("Company Inventory", m.buildRows(m.state.CompanyInventory))
	player := renderSection("Player Inventory", m.buildRows(m.state.PlayerInventory))
	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("esc/q: back")
	content := lipgloss.JoinVertical(lipgloss.Left, company, "", player, "", footer)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func renderSection(title string, rows []row) string {
	lines := []string{title, ""}
	if len(rows) == 0 {
		lines = append(lines, "  (empty)")
	} else {
		for _, r := range rows {
			status := "OK"
			if r.broken {
				status = "Broken"
			}
			lines = append(lines, fmt.Sprintf("  %s x%d  durability:%d  status:%s", r.name, r.quantity, r.durability, status))
		}
	}
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(86).Padding(1, 2).Render(strings.Join(lines, "\n"))
}

func (m Model) buildRows(inv domain.Inventory) []row {
	entries := inv.Entries()
	rows := make([]row, 0, len(entries))
	for _, e := range entries {
		item, ok := m.catalog.Item(e.ItemID)
		name := string(e.ItemID)
		if ok {
			name = item.DisplayName
		}
		rows = append(rows, row{
			name:       name,
			quantity:   e.Quantity,
			durability: e.RemainingDurabilityDays,
			broken:     e.RemainingDurabilityDays <= 0,
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].broken != rows[j].broken {
			return !rows[i].broken
		}
		if rows[i].name != rows[j].name {
			return rows[i].name < rows[j].name
		}
		if rows[i].durability != rows[j].durability {
			return rows[i].durability > rows[j].durability
		}
		return rows[i].quantity > rows[j].quantity
	})
	return rows
}
