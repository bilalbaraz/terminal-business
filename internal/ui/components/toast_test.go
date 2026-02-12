package components

import "testing"

func TestToastShowHide(t *testing.T) {
	var toast Toast
	toast.Show("ok")
	if !toast.Visible || toast.Message != "ok" {
		t.Fatal("toast not shown")
	}
	toast.Hide()
	if toast.Visible {
		t.Fatal("toast not hidden")
	}
}
