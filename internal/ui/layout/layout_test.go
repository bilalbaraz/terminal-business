package layout

import "testing"

func TestComputeBreakpointsAndBounds(t *testing.T) {
	r := Compute(120, 40)
	if r.Mode != WideMode {
		t.Fatalf("got mode %s", r.Mode)
	}
	if r.ScrollableRows <= 0 {
		t.Fatal("expected rows")
	}
	c := Compute(70, 10)
	if c.Mode != CompactMode {
		t.Fatalf("got mode %s", c.Mode)
	}
	if c.SidebarWidth != c.Width-4 || c.MainWidth != c.Width-4 {
		t.Fatalf("unexpected compact widths %+v", c)
	}
	tiny := Compute(1, 1)
	if tiny.Width < 20 || tiny.Height < 8 {
		t.Fatalf("expected clamped tiny %+v", tiny)
	}
}

func TestClipRows(t *testing.T) {
	in := []int{1, 2, 3, 4}
	if len(ClipRows(in, 0)) != 4 {
		t.Fatal("expected unchanged on non-positive")
	}
	if len(ClipRows(in, 2)) != 2 {
		t.Fatal("expected clipped")
	}
}
