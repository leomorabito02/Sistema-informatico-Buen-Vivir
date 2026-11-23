package storage

import (
	"encoding/base64"
	"testing"
)

func TestResolveStorageClientOptions_Emulator(t *testing.T) {
	// Arrange
	emulatorHost := "http://localhost:4443"

	// Act
	opts, err := resolveStorageClientOptions(emulatorHost, "")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(opts) != 2 {
		t.Fatalf("expected 2 emulator options, got %d", len(opts))
	}
}

func TestResolveStorageClientOptions_Base64EnvFallback(t *testing.T) {
	// Arrange
	rawJSON := `{"type":"service_account","project_id":"ceibuenvivir"}`
	encoded := base64.StdEncoding.EncodeToString([]byte(rawJSON))
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS_BASE64", "")
	t.Setenv("FIREBASE_CREDENTIALS_BASE64", encoded)

	// Act
	opts, err := resolveStorageClientOptions("", "")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(opts) != 1 {
		t.Fatalf("expected 1 credentials option, got %d", len(opts))
	}
}

func TestResolveStorageClientOptions_InvalidBase64(t *testing.T) {
	// Arrange
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS_BASE64", "not-valid-base64!!!")
	t.Setenv("FIREBASE_CREDENTIALS_BASE64", "")

	// Act
	_, err := resolveStorageClientOptions("", "")

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid base64 credentials")
	}
}

func TestResolveGoogleClientOptions_Base64OverridesLegacyPath(t *testing.T) {
	// Arrange
	rawJSON := `{"type":"service_account","project_id":"ceibuenvivir"}`
	encoded := base64.StdEncoding.EncodeToString([]byte(rawJSON))
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS_BASE64", encoded)
	t.Setenv("FIREBASE_CREDENTIALS_BASE64", "")
	t.Setenv("FIREBASE_CREDENTIALS_JSON", "")
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "./nonexistent-service-account.json")

	// Act
	opts, err := ResolveGoogleClientOptions("./nonexistent-service-account.json")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(opts) != 1 {
		t.Fatalf("expected 1 credentials option, got %d", len(opts))
	}
}

func TestResolveGoogleClientOptions_EmptyFallsBackToADC(t *testing.T) {
	// Arrange
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS_BASE64", "")
	t.Setenv("FIREBASE_CREDENTIALS_BASE64", "")
	t.Setenv("FIREBASE_CREDENTIALS_JSON", "")
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "")

	// Act
	opts, err := ResolveGoogleClientOptions("")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if opts != nil {
		t.Fatalf("expected nil options for ADC fallback, got %d", len(opts))
	}
}

func TestResolveFirestoreDatabase(t *testing.T) {
	for _, tc := range []struct {
		name string
		val  string
		want string
	}{
		{"unset or empty defaults to default db", "", ""},
		{"whitespace defaults to default db", "   ", ""},
		{"explicit default marker", "(default)", ""},
		{"literal null string", "null", ""},
		{"uppercase NULL string", "NULL", ""},
		{"custom database name", "buenvivir-prod", "buenvivir-prod"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("FIRESTORE_DATABASE", tc.val)
			if got := resolveFirestoreDatabase(); got != tc.want {
				t.Fatalf("resolveFirestoreDatabase() = %q, want %q", got, tc.want)
			}
		})
	}
}
