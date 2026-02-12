package components

type Toast struct {
	Message string
	Visible bool
}

func (t *Toast) Show(msg string) {
	t.Message = msg
	t.Visible = true
}

func (t *Toast) Hide() {
	t.Visible = false
}
