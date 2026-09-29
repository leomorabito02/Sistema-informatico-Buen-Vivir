package services

import "testing"

func TestIsApprovedStatus(t *testing.T) {
	cases := map[string]bool{
		"approved": true,
		"accepted": true, // valor legado, filas reales en Neon
		"pending":  false,
		"rejected": false,
		"":         false,
	}
	for status, want := range cases {
		if got := isApprovedStatus(status); got != want {
			t.Errorf("isApprovedStatus(%q) = %v, want %v", status, got, want)
		}
	}
}
