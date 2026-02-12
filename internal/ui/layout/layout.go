package layout

type Mode string

const (
	WideMode    Mode = "wide"
	CompactMode Mode = "compact"
)

type Regions struct {
	Width          int
	Height         int
	Mode           Mode
	SidebarWidth   int
	MainWidth      int
	RightWidth     int
	ScrollableRows int
}

func Compute(width, height int) Regions {
	if width < 20 {
		width = 20
	}
	if height < 8 {
		height = 8
	}
	mode := WideMode
	if width < 90 {
		mode = CompactMode
	}
	sidebar := width / 4
	if sidebar < 18 {
		sidebar = 18
	}
	if sidebar > 30 {
		sidebar = 30
	}
	right := width / 4
	if right < 20 {
		right = 20
	}
	if right > 30 {
		right = 30
	}
	main := width - sidebar - right - 4
	if main < 24 {
		main = 24
	}
	rows := height - 8
	if rows < 3 {
		rows = 3
	}
	if mode == CompactMode {
		sidebar = width - 4
		main = width - 4
		right = width - 4
	}
	return Regions{
		Width:          width,
		Height:         height,
		Mode:           mode,
		SidebarWidth:   sidebar,
		MainWidth:      main,
		RightWidth:     right,
		ScrollableRows: rows,
	}
}

func ClipRows[T any](rows []T, max int) []T {
	if max <= 0 || len(rows) <= max {
		return rows
	}
	return rows[:max]
}
