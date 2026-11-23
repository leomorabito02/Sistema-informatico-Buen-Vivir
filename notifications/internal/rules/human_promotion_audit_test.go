package rules_test

import (
	"context"
	"strings"
	"testing"

	"buenvivir/notifications/internal/events"
	"buenvivir/notifications/internal/rules"
)

// rutasDelFront son las pantallas que existen en frontend/src/app/app.routes.ts
// a las que pueden apuntar las notificaciones. Si una regla manda a otra
// ruta, el usuario cae en la página 404.
var rutasDelFront = []string{
	"/dashboard/promocion-humana/participantes/",
	"/dashboard/promocion-humana",
	"/dashboard/mi-proceso",
	"/dashboard/cursos/",
	"/dashboard/cursos",
	"/dashboard/contenido-docente",
	"/dashboard/mis-contenidos",
	"/dashboard/mi-perfil",
	"/dashboard/usuarios",
	"/dashboard/ofertas",
	"/dashboard/campanas",
	"/campanas-publico",
}

func rutaExiste(u string) bool {
	for _, r := range rutasDelFront {
		if u == r || (strings.HasSuffix(r, "/") && strings.HasPrefix(u, r) && len(u) > len(r)) {
			return true
		}
	}
	return false
}

// Las notificaciones de Promoción Humana apuntan a la ficha del participante.
func TestPromocionHumana_EnlacesALaFichaDelParticipante(t *testing.T) {
	ctx := context.Background()
	ficha := events.FichaPayload{FichaID: "f-1", ParticipanteID: "part-9", CreadoPor: "vol-1"}
	hito := events.HitoPayload{HitoID: "h-1", FichaID: "f-1", ParticipanteID: "part-9", ResponsableID: "vol-2", VisibleParticipante: true}

	casos := []struct {
		rule    rules.Rule
		payload any
		actor   string
		planes  int
	}{
		{&rules.FichaCreadaRule{}, ficha, "vol-1", 1},
		{&rules.FichaActualizadaRule{}, ficha, "admin-1", 1}, // editó otra persona: avisa al creador
		{&rules.FichaActualizadaRule{}, ficha, "vol-1", 0},   // editó el creador: no avisa
		{&rules.HitoRegistradoRule{}, hito, "vol-1", 2},      // participante + responsable
		{&rules.HitoActualizadoRule{}, hito, "vol-1", 1},
	}
	for _, c := range casos {
		env := newTestEnvelope(t, c.rule.EventType(), c.payload, c.actor)
		planes, err := c.rule.Build(ctx, env)
		if err != nil {
			t.Fatalf("%s: %v", c.rule.EventType(), err)
		}
		if len(planes) != c.planes {
			t.Fatalf("%s (actor %s): %d planes, esperaba %d", c.rule.EventType(), c.actor, len(planes), c.planes)
		}
		for _, p := range planes {
			if p.ActionURL == nil || !rutaExiste(*p.ActionURL) {
				t.Errorf("%s: enlace %v no existe en el frontend", c.rule.EventType(), p.ActionURL)
			}
			if p.DedupeKey == "" {
				t.Errorf("%s: sin dedupe_key, un reintento duplicaría la notificación", c.rule.EventType())
			}
		}
	}

	planes, _ := (&rules.FichaCreadaRule{}).Build(ctx, newTestEnvelope(t, events.FichaCreada, ficha, "vol-1"))
	if *planes[0].ActionURL != "/dashboard/promocion-humana/participantes/part-9" {
		t.Errorf("ficha.creada debería llevar a la ficha del participante, lleva a %s", *planes[0].ActionURL)
	}
}

// Ningún enlace de ninguna regla puede apuntar a una ruta inexistente
// (también estaban rotos /dashboard/educacion/... y /dashboard/perfil).
func TestTodasLasReglas_EnlacesExistenEnElFront(t *testing.T) {
	ctx := context.Background()
	payloads := map[string]any{
		events.InscripcionCreada:           events.InscripcionCreadaPayload{InscripcionID: "i", CursoID: "c", ParticipanteID: "p", DocenteID: "d"},
		events.ContenidoDocenteSubido:      events.ContenidoDocenteSubidoPayload{DocenteID: "d", CursoID: "c"},
		events.ContenidoDocenteModerado:    events.ContenidoDocenteModeradoPayload{DocenteID: "d", CursoID: "c", Resultado: "aprobado"},
		events.CursoCompletadoParticipante: events.CursoCompletadoPayload{CursoID: "c", ParticipanteID: "p"},
		events.CursoCreado:                 events.CursoPayload{CursoID: "c"},
		events.CursoEliminado:              events.CursoPayload{CursoID: "c"},
	}
	reglas := []rules.Rule{
		&rules.InscripcionCreadaRule{}, &rules.ContenidoDocenteSubidoRule{}, &rules.ContenidoDocenteModeradoRule{},
		&rules.CursoCompletadoRule{}, &rules.CursoCreadoRule{}, &rules.CursoEliminadoRule{},
	}
	for _, r := range reglas {
		p, ok := payloads[r.EventType()]
		if !ok {
			continue
		}
		planes, err := r.Build(ctx, newTestEnvelope(t, r.EventType(), p, "actor-1"))
		if err != nil {
			t.Fatalf("%s: %v", r.EventType(), err)
		}
		for _, pl := range planes {
			if pl.ActionURL != nil && !rutaExiste(*pl.ActionURL) {
				t.Errorf("%s: enlace %s no existe en el frontend", r.EventType(), *pl.ActionURL)
			}
		}
	}
}
