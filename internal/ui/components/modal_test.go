package components

import "testing"

func TestModalSelection(t *testing.T) {
	m := Modal{Actions: []string{"A", "B"}}
	if m.Selected() != "A" {
		t.Fatal("expected A")
	}
	m.Right()
	if m.Selected() != "B" {
		t.Fatal("expected B")
	}
	m.Right()
	if m.Selected() != "A" {
		t.Fatal("expected wrap to A")
	}
	m.Left()
	if m.Selected() != "B" {
		t.Fatal("expected wrap to B")
	}
}

func TestModalEmptyAndOutOfRange(t *testing.T) {
	m := Modal{}
	m.Left()
	m.Right()
	if m.Selected() != "" {
		t.Fatal("expected empty selected")
	}
	m = Modal{Actions: []string{"A"}, SelectedIndex: 9}
	if m.Selected() != "" {
		t.Fatal("expected out of range selected to be empty")
	}
}
