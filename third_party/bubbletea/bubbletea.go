package bubbletea

import "time"

type Msg interface{}

type Cmd func() Msg

type Model interface {
	Init() Cmd
	Update(Msg) (Model, Cmd)
	View() string
}

type Program struct {
	model Model
}

type ProgramOption func(*Program)

func NewProgram(model Model, _ ...ProgramOption) *Program {
	return &Program{model: model}
}

func WithAltScreen() ProgramOption { return func(*Program) {} }

func (p *Program) Run() (Model, error) {
	if p.model == nil {
		return nil, nil
	}
	if cmd := p.model.Init(); cmd != nil {
		_, _ = p.model.Update(cmd())
	}
	return p.model, nil
}

func Batch(cmds ...Cmd) Cmd {
	return func() Msg {
		var last Msg
		for _, cmd := range cmds {
			if cmd == nil {
				continue
			}
			last = cmd()
		}
		return last
	}
}

func Tick(_ time.Duration, fn func(time.Time) Msg) Cmd {
	return func() Msg {
		return fn(time.Now())
	}
}

type quitMsg struct{}

var Quit Cmd = func() Msg { return quitMsg{} }

type WindowSizeMsg struct {
	Width  int
	Height int
}

type KeyType int

const (
	KeyRunes KeyType = iota
	KeyEnter
	KeyEsc
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyBackspace
	KeyCtrlC
)

type KeyMsg struct {
	Type  KeyType
	Runes []rune
}

func (k KeyMsg) String() string {
	switch k.Type {
	case KeyEnter:
		return "enter"
	case KeyEsc:
		return "esc"
	case KeyUp:
		return "up"
	case KeyDown:
		return "down"
	case KeyLeft:
		return "left"
	case KeyRight:
		return "right"
	case KeyBackspace:
		return "backspace"
	case KeyCtrlC:
		return "ctrl+c"
	case KeyRunes:
		if len(k.Runes) == 1 {
			return string(k.Runes[0])
		}
		return string(k.Runes)
	default:
		return ""
	}
}
