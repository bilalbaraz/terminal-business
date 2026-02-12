package components

type MenuList struct {
	Items  []string
	Cursor int
}

func NewMenuList(items []string) MenuList {
	return MenuList{Items: items}
}

func (m *MenuList) MoveUp() {
	if len(m.Items) == 0 {
		return
	}
	m.Cursor--
	if m.Cursor < 0 {
		m.Cursor = len(m.Items) - 1
	}
}

func (m *MenuList) MoveDown() {
	if len(m.Items) == 0 {
		return
	}
	m.Cursor++
	if m.Cursor >= len(m.Items) {
		m.Cursor = 0
	}
}

func (m MenuList) Selected() string {
	if len(m.Items) == 0 {
		return ""
	}
	if m.Cursor < 0 || m.Cursor >= len(m.Items) {
		return ""
	}
	return m.Items[m.Cursor]
}
