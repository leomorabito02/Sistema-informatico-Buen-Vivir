package events

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

var (
	pushSubLine = regexp.MustCompile(`(?m)^ensure_push_sub ([a-z0-9-]+) .*\\\n\s+"\$\(filter ([^)]*)\)"`)
)

// InfraFilters lee infra/pubsub/setup.sh y devuelve los tipos de evento del filtro de
// cada suscripción. Los tests lo usan para comprobar que coincide con lo que procesa
// el consumidor.
func InfraFilters(setupPath string) (map[string][]string, error) {
	raw, err := os.ReadFile(setupPath)
	if err != nil {
		return nil, err
	}
	out := map[string][]string{}
	for _, m := range pushSubLine.FindAllStringSubmatch(string(raw), -1) {
		types := strings.Fields(m[2])
		sort.Strings(types)
		out[m[1]] = types
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no se encontraron suscripciones en %s", setupPath)
	}
	return out, nil
}

// SameTypes compara dos listas de tipos sin importar el orden.
func SameTypes(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	x := append([]string(nil), a...)
	y := append([]string(nil), b...)
	sort.Strings(x)
	sort.Strings(y)
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}
