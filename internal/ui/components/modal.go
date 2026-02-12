package components

type Modal struct {
	Open          bool
	Title         string
	Body          string
	Actions       []string
	SelectedIndex int
}

func (m *Modal) Left() {
	if len(m.Actions) == 0 {
		return
	}
	m.SelectedIndex--
	if m.SelectedIndex < 0 {
		m.SelectedIndex = len(m.Actions) - 1
	}
}

func (m *Modal) Right() {
	if len(m.Actions) == 0 {
		return
	}
	m.SelectedIndex++
	if m.SelectedIndex >= len(m.Actions) {
		m.SelectedIndex = 0
	}
}

func (m Modal) Selected() string {
	if len(m.Actions) == 0 || m.SelectedIndex < 0 || m.SelectedIndex >= len(m.Actions) {
		return ""
	}
	return m.Actions[m.SelectedIndex]
}
