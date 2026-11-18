package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env.test")

	content := `
# Comment line
KEY_PLAIN=plain_value
KEY_DOUBLE="double_quoted"
KEY_SINGLE='single_quoted'
KEY_EMPTY=
INVALID_LINE_NO_EQUALS
=NO_KEY
KEY_EXISTING=new_value
`
	err := os.WriteFile(envPath, []byte(content), 0600)
	assert.NoError(t, err)

	// Pre-set existing key to test precedence
	t.Setenv("KEY_EXISTING", "original_value")

	// Call LoadEnvFile with the test file
	LoadEnvFile(envPath)

	assert.Equal(t, "plain_value", os.Getenv("KEY_PLAIN"))
	assert.Equal(t, "double_quoted", os.Getenv("KEY_DOUBLE"))
	assert.Equal(t, "single_quoted", os.Getenv("KEY_SINGLE"))
	assert.Equal(t, "original_value", os.Getenv("KEY_EXISTING"))

	// Non-existent path silently ignored
	LoadEnvFile(filepath.Join(tmpDir, "does_not_exist"))

	// Default paths (should not crash)
	LoadEnvFile()
}
