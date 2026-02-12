package components

import (
	"strings"
	"testing"
)

func TestErrorScreen(t *testing.T) {
	view := ErrorScreen("Fatal", "boom")
	if !strings.Contains(view, "Fatal") || !strings.Contains(view, "boom") {
		t.Fatalf("unexpected view: %s", view)
	}
}
