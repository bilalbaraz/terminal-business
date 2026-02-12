package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type spinnerTickMsg struct{}

type Spinner struct {
	frames []string
	idx    int
}

func NewSpinner() Spinner {
	return Spinner{frames: []string{"|", "/", "-", "\\"}}
}

func (s *Spinner) Tick() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(time.Time) tea.Msg { return spinnerTickMsg{} })
}

func (s *Spinner) Update(msg tea.Msg) {
	if _, ok := msg.(spinnerTickMsg); ok {
		s.idx = (s.idx + 1) % len(s.frames)
	}
}

func (s Spinner) View() string {
	return s.frames[s.idx]
}
