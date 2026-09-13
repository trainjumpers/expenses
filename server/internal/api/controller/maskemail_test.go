package controller

import "testing"

func TestMaskEmail(t *testing.T) {
	cases := map[string]string{
		"user@example.com": "***@example.com",
		"nodomain":         "***",
		"":                 "***",
		"@example.com":     "***",
	}
	for input, expected := range cases {
		if got := maskEmail(input); got != expected {
			t.Fatalf("maskEmail(%q) = %q, want %q", input, got, expected)
		}
	}
}
