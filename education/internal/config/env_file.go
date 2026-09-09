package config

import (
	"bufio"
	"os"
	"strings"
)

// loadEnvFile es un parser mínimo de archivos ".env" (KEY=VALUE por
// línea, comentarios con "#", comillas simples/dobles opcionales) para
// no depender de una librería externa como godotenv.
//
// Las variables que ya existen en el entorno real del proceso (por
// ejemplo, las inyectadas por "env_file" de docker-compose, o por la
// plataforma en producción) SIEMPRE tienen prioridad: el archivo solo
// completa lo que falte. Esto permite que el mismo binario funcione
// igual corriendo suelto en desarrollo (leyendo el archivo) o dentro
// de un contenedor (con las variables ya en el entorno).
//
// Si el archivo no existe, no es un error: simplemente no hay nada
// que completar (caso típico en producción, donde las variables las
// provee la plataforma).
func loadEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = unquote(value)

		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, value)
	}

	return scanner.Err()
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
