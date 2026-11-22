package services

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"buenvivir/education/internal/models"
)

// UserNamesSource devuelve user_id -> nombre (ver AuthServiceClient.GetUserNames).
type UserNamesSource func(ctx context.Context, userIDs []string) (map[string]string, error)

// TeacherNames completa Course.ResponsibleTeacherName antes de responder
// (D3, 22/11/2026: en "Mis cursos" se veía "Docente: usr-…"). Guarda los
// nombres unos minutos para no consultar a auth en cada listado. Si auth
// no responde, el curso sale sin nombre y el front muestra el texto
// genérico; nunca falla el pedido por esto.
type TeacherNames struct {
	source  UserNamesSource
	ttl     time.Duration
	timeout time.Duration
	now     func() time.Time

	mu    sync.Mutex
	cache map[string]nombreCacheado
}

type nombreCacheado struct {
	nombre string
	vence  time.Time
}

// NewTeacherNames arma el completador. source nil = no hace nada.
func NewTeacherNames(source UserNamesSource) *TeacherNames {
	return &TeacherNames{
		source:  source,
		ttl:     10 * time.Minute,
		timeout: 3 * time.Second,
		now:     time.Now,
		cache:   map[string]nombreCacheado{},
	}
}

// Fill completa el nombre del docente en cada curso de la lista.
func (t *TeacherNames) Fill(ctx context.Context, courses []models.Course) {
	if t == nil || t.source == nil || len(courses) == 0 {
		return
	}
	ids := make([]string, 0, len(courses))
	for _, c := range courses {
		if c.ResponsibleTeacher != "" {
			ids = append(ids, c.ResponsibleTeacher)
		}
	}
	nombres := t.resolver(ctx, ids)
	for i := range courses {
		if n, ok := nombres[courses[i].ResponsibleTeacher]; ok {
			courses[i].ResponsibleTeacherName = n
		}
	}
}

// FillOne completa el nombre del docente de un curso.
func (t *TeacherNames) FillOne(ctx context.Context, course *models.Course) {
	if course == nil {
		return
	}
	lista := []models.Course{*course}
	t.Fill(ctx, lista)
	course.ResponsibleTeacherName = lista[0].ResponsibleTeacherName
}

func (t *TeacherNames) resolver(ctx context.Context, ids []string) map[string]string {
	out := map[string]string{}
	var faltan []string
	vistos := map[string]bool{}
	ahora := t.now()
	t.mu.Lock()
	for _, id := range ids {
		if vistos[id] {
			continue
		}
		vistos[id] = true
		if c, ok := t.cache[id]; ok && ahora.Before(c.vence) {
			if c.nombre != "" {
				out[id] = c.nombre
			}
			continue
		}
		faltan = append(faltan, id)
	}
	t.mu.Unlock()
	if len(faltan) == 0 {
		return out
	}

	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()
	nuevos, err := t.source(ctx, faltan)
	if err != nil {
		slog.Warn("no se pudieron resolver los nombres de docentes", slog.String("error", err.Error()))
		if len(nuevos) == 0 {
			return out
		}
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	vence := t.now().Add(t.ttl)
	for _, id := range faltan {
		n := nuevos[id]
		// Un id sin nombre también se guarda (vacío) para no preguntarlo
		// en cada pedido, salvo que la consulta haya fallado.
		if n != "" || err == nil {
			t.cache[id] = nombreCacheado{nombre: n, vence: vence}
		}
		if n != "" {
			out[id] = n
		}
	}
	return out
}
