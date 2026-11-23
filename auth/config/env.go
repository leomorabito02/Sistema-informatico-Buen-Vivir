package config

import (
	"bufio"
	"os"
	"strings"
)

// LoadEnvFile lee un archivo .env local sin pisar las variables que ya existen en el
// entorno. Si el archivo no existe, sigue sin error.
func LoadEnvFile(paths ...string) {
	if len(paths) == 0 {
		paths = []string{".env", "../.env"}
	}

	for _, path := range paths {
		parseFile(path)
	}
}

func parseFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parseLine(scanner.Text())
	}
	_ = scanner.Err()
}

func parseLine(rawLine string) {
	line := strings.TrimSpace(rawLine)
	if line == "" || strings.HasPrefix(line, "#") {
		return
	}

	key, value, found := strings.Cut(line, "=")
	if !found {
		return
	}

	key = strings.TrimSpace(key)
	value = unquote(strings.TrimSpace(value))

	if key == "" {
		return
	}
	if _, exists := os.LookupEnv(key); exists {
		return
	}
	_ = os.Setenv(key, value)
}

func unquote(value string) string {
	if len(value) >= 2 {
		first, last := value[0], value[len(value)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}
