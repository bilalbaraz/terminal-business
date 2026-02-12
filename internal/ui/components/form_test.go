package components

import "testing"

func TestValidateCompanyName(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"", "Company name is required."},
		{"x", "Use 2-40 characters."},
		{"a@b", "Only letters, numbers, spaces, -, ', & are allowed."},
		{"Acme Labs", ""},
	}
	for _, tc := range cases {
		if got := ValidateCompanyName(tc.name); got != tc.want {
			t.Fatalf("name %q got %q want %q", tc.name, got, tc.want)
		}
	}
}
