package config

import "testing"

func TestResolveFirestoreDatabase(t *testing.T) {
	for _, tc := range []struct {
		name string
		val  string
		want string
	}{
		{"empty", "", ""},
		{"default", "(default)", ""},
		{"null", "null", ""},
		{"custom", "buenvivir-prod", "buenvivir-prod"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("FIRESTORE_DATABASE", tc.val)
			if got := resolveFirestoreDatabase(); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
