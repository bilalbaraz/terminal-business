package components

import "testing"

func TestMenuListWrapNavigation(t *testing.T) {
	m := NewMenuList([]string{"a", "b", "c"})
	if got := m.Selected(); got != "a" {
		t.Fatalf("got %q", got)
	}
	m.MoveUp()
	if got := m.Selected(); got != "c" {
		t.Fatalf("got %q", got)
	}
	m.MoveDown()
	if got := m.Selected(); got != "a" {
		t.Fatalf("got %q", got)
	}
}

func TestMenuListEmptyAndOutOfRange(t *testing.T) {
	m := NewMenuList(nil)
	m.MoveDown()
	m.MoveUp()
	if got := m.Selected(); got != "" {
		t.Fatalf("got %q", got)
	}
	m = NewMenuList([]string{"x"})
	m.Cursor = 2
	if got := m.Selected(); got != "" {
		t.Fatalf("got %q", got)
	}
}
