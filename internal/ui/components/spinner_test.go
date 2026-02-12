package components

import "testing"

func TestSpinnerTick(t *testing.T) {
	s := NewSpinner()
	first := s.View()
	msg := s.Tick()()
	s.Update(msg)
	if s.View() == first {
		t.Fatal("expected spinner frame to advance")
	}
}
