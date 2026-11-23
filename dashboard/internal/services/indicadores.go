package services

import (
	"fmt"
	"time"

	"buenvivir/dashboard/internal/models"
)

// Cálculo de los indicadores y las gráficas del dashboard (HU-52 a HU-54). Funciones
// puras, sin red.

// Zona es la hora del centro (Argentina, UTC-3 sin horario de verano).
// Los cortes de día y de mes se calculan en esta zona, no en UTC: un alta
// del 31/03 a las 22 h es de marzo.
var Zona = time.FixedZone("ART", -3*60*60)

// MaxMesesPeriodo acota el período que se puede pedir (diez años).
const MaxMesesPeriodo = 120

// Módulos del filtro de HU-54.
const (
	ModuloTodos     = "todos"
	ModuloUsuarios  = "usuarios"
	ModuloEducacion = "educacion"
	ModuloPromocion = "promocion"
	ModuloInsercion = "insercion"
)

// Filtros del dashboard. Desde y Hasta son días (a las 00:00 en Zona);
// AreaID vacío = todas las áreas formativas.
type Filtros struct {
	Desde  time.Time
	Hasta  time.Time
	Modulo string
	AreaID string
}

// Datos crudos de cada fuente. Un puntero nil significa que ese servicio
// no respondió: sus indicadores salen "No disponible" y el resto se
// calcula igual.
type Datos struct {
	Participantes *[]models.Participante
	Cursos        *[]models.Curso
	Habilidades   *[]models.ParticipanteHabilidad
	Postulaciones *[]models.Postulacion
	Promocion     *models.Promocion
	Equipo        *models.Equipo
}

type metaIndicador struct {
	id, titulo, descripcion, modulo, fuente string
}

// Indicadores en el orden en que se muestran: los cinco de HU-52 y los
// dos de Promoción Humana.
var metaIndicadores = []metaIndicador{
	{"participantesActivos", "Participantes activos", "Participantes con la cuenta activa al cierre del período.", ModuloUsuarios, "autenticacion"},
	{"empresasActivas", "Empresas activas", "Empresas con la cuenta activa y aprobada hoy.", ModuloUsuarios, "autenticacion"},
	{"docentesActivos", "Docentes activos", "Docentes con la cuenta activa y aprobada hoy.", ModuloUsuarios, "autenticacion"},
	{"voluntariosActivos", "Voluntarios activos", "Voluntarios con la cuenta activa y aprobada hoy.", ModuloUsuarios, "autenticacion"},
	{"cursosEnDesarrollo", "Cursos en desarrollo", "Cursos publicados con actividades pendientes.", ModuloEducacion, "educacion"},
	{"cursosFinalizados", "Cursos finalizados", "Cursos publicados cuyas actividades ya ocurrieron todas.", ModuloEducacion, "educacion"},
	{"habilidadesDesarrolladas", "Habilidades desarrolladas", "Habilidades aceptadas por la administración en el período.", ModuloEducacion, "educacion"},
	{"insercionesLaborales", "Inserciones laborales", "Postulaciones que pasaron a «seleccionado» en el período.", ModuloInsercion, "insercion"},
	{"fichasActivas", "Fichas activas", "Fichas de proceso personal abiertas al cierre del período.", ModuloPromocion, "promocion"},
	{"participantesConSeguimiento", "Participantes con seguimiento", "Participantes con al menos un hito de avance en el período.", ModuloPromocion, "promocion"},
}

// ModuloValido dice si m es uno de los valores del filtro.
func ModuloValido(m string) bool {
	switch m {
	case ModuloTodos, ModuloUsuarios, ModuloEducacion, ModuloPromocion, ModuloInsercion:
		return true
	}
	return false
}

// ── Reglas de negocio ──────────────────────────────────────────────────

const (
	cursoFuera        = "fuera"
	cursoEnDesarrollo = "en_desarrollo"
	cursoFinalizado   = "finalizado"
)

// EstadoCurso: a la fecha ref, un curso
//   - no cuenta ("fuera") si es un borrador o todavía no existía;
//   - está "finalizado" si tiene actividades y todas ya ocurrieron;
//   - está "en_desarrollo" si tiene alguna actividad pendiente, o si
//     todavía no tiene actividades cargadas (está activo y en preparación).
func EstadoCurso(c models.Curso, ref time.Time) string {
	if c.Estado != "activo" || c.FechaCreacion.After(ref) {
		return cursoFuera
	}
	if len(c.Actividades) == 0 {
		return cursoEnDesarrollo
	}
	for _, a := range c.Actividades {
		if a.Fecha.After(ref) {
			return cursoEnDesarrollo
		}
	}
	return cursoFinalizado
}

// ParticipanteActivoEn: ya estaba dado de alta y no tenía baja lógica a la
// fecha ref. Una baja sin fecha no cuenta (no se sabe desde cuándo).
func ParticipanteActivoEn(p models.Participante, ref time.Time) bool {
	if p.Rol != "participante" || p.FechaAlta.After(ref) {
		return false
	}
	if p.Estado == "baja" {
		return p.FechaBaja != nil && p.FechaBaja.After(ref)
	}
	return true
}

// FichaActivaEn: mismo criterio que ParticipanteActivoEn.
func FichaActivaEn(f models.FichaPromocion, ref time.Time) bool {
	if f.FechaAlta.After(ref) {
		return false
	}
	if f.Estado == "baja" {
		return f.FechaBaja != nil && f.FechaBaja.After(ref)
	}
	return true
}

// ParticipantesConSeguimiento cuenta participantes distintos con algún
// hito entre desde y hasta (inclusive). Los hitos son AAAA-MM-DD sin
// hora: se comparan como texto para que la zona horaria no corra el día.
func ParticipantesConSeguimiento(hitos []models.HitoPromocion, desde, hasta time.Time) int {
	d, h := fechaTexto(desde), fechaTexto(hasta)
	vistos := map[string]bool{}
	for _, x := range hitos {
		if x.Fecha >= d && x.Fecha <= h {
			vistos[x.ParticipanteID] = true
		}
	}
	return len(vistos)
}

// ── Fechas ─────────────────────────────────────────────────────────────

func inicioDelDia(t time.Time) time.Time {
	t = t.In(Zona)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, Zona)
}

func finDelDia(t time.Time) time.Time {
	return inicioDelDia(t).AddDate(0, 0, 1).Add(-time.Nanosecond)
}

func fechaTexto(t time.Time) string { return t.In(Zona).Format("2006-01-02") }

func enRango(t, desde, hasta time.Time) bool {
	return !t.Before(inicioDelDia(desde)) && !t.After(finDelDia(hasta))
}

// Mes del período: Inicio es el primer instante que cuenta y Corte el
// último (fin de mes, o el fin de Hasta en el último mes).
type Mes struct {
	Clave  string // "AAAA-MM"
	Inicio time.Time
	Corte  time.Time
}

// MesesDelPeriodo devuelve un mes por punto de las gráficas, sin mirar
// más allá del filtro.
func MesesDelPeriodo(desde, hasta time.Time) []Mes {
	var out []Mes
	ini, fin := inicioDelDia(desde), finDelDia(hasta)
	if ini.After(fin) {
		return out
	}
	for cursor := time.Date(ini.Year(), ini.Month(), 1, 0, 0, 0, 0, Zona); !cursor.After(fin); cursor = cursor.AddDate(0, 1, 0) {
		ultimo := cursor.AddDate(0, 1, 0).Add(-time.Nanosecond)
		m := Mes{Clave: fmt.Sprintf("%04d-%02d", cursor.Year(), cursor.Month()), Inicio: cursor, Corte: ultimo}
		if m.Inicio.Before(ini) {
			m.Inicio = ini
		}
		if m.Corte.After(fin) {
			m.Corte = fin
		}
		out = append(out, m)
	}
	return out
}

// MesesEntre cuenta los meses calendario que toca el período.
func MesesEntre(desde, hasta time.Time) int {
	d, h := desde.In(Zona), hasta.In(Zona)
	return (h.Year()-d.Year())*12 + int(h.Month()-d.Month()) + 1
}

// ── Cálculo ────────────────────────────────────────────────────────────

func ptr(n int) *int { return &n }

func contar[T any](items []T, ok func(T) bool) int {
	n := 0
	for _, i := range items {
		if ok(i) {
			n++
		}
	}
	return n
}

func filtrar[T any](items []T, ok func(T) bool) []T {
	out := []T{}
	for _, i := range items {
		if ok(i) {
			out = append(out, i)
		}
	}
	return out
}

// CalcularResumen arma los indicadores y las gráficas para los filtros.
func CalcularResumen(d Datos, f Filtros) models.Resumen {
	corte := finDelDia(f.Hasta)
	meses := MesesDelPeriodo(f.Desde, f.Hasta)

	porMes := func(fn func(m Mes) int) []models.PuntoSerie {
		out := make([]models.PuntoSerie, len(meses))
		for i, m := range meses {
			out[i] = models.PuntoSerie{Mes: m.Clave, Valor: fn(m)}
		}
		return out
	}

	valores := map[string]*int{}
	series := map[string][]models.SerieGrafica{}

	if d.Participantes != nil {
		ps := *d.Participantes
		activos := func(ref time.Time) int {
			return contar(ps, func(p models.Participante) bool { return ParticipanteActivoEn(p, ref) })
		}
		valores["participantesActivos"] = ptr(activos(corte))
		series["participantes"] = []models.SerieGrafica{
			{Nombre: "Participantes activos", Puntos: porMes(func(m Mes) int { return activos(m.Corte) })},
		}
	}

	// Cantidades de hoy (no dependen del período elegido).
	if d.Equipo != nil {
		valores["empresasActivas"] = ptr(d.Equipo.Empresas)
		valores["docentesActivos"] = ptr(d.Equipo.Docentes)
		valores["voluntariosActivos"] = ptr(d.Equipo.Voluntarios)
	}

	// El área formativa afecta solo a cursos y habilidades.
	if d.Cursos != nil {
		cursos := filtrar(*d.Cursos, func(c models.Curso) bool { return f.AreaID == "" || c.AreaID == f.AreaID })
		en := func(ref time.Time, estado string) int {
			return contar(cursos, func(c models.Curso) bool { return EstadoCurso(c, ref) == estado })
		}
		valores["cursosEnDesarrollo"] = ptr(en(corte, cursoEnDesarrollo))
		valores["cursosFinalizados"] = ptr(en(corte, cursoFinalizado))
		series["cursos"] = []models.SerieGrafica{
			{Nombre: "En desarrollo", Puntos: porMes(func(m Mes) int { return en(m.Corte, cursoEnDesarrollo) })},
			{Nombre: "Finalizados", Puntos: porMes(func(m Mes) int { return en(m.Corte, cursoFinalizado) })},
		}
	}

	if d.Habilidades != nil {
		hs := filtrar(*d.Habilidades, func(h models.ParticipanteHabilidad) bool {
			return h.Estado == "aceptada" && (f.AreaID == "" || h.AreaID == f.AreaID)
		})
		valores["habilidadesDesarrolladas"] = ptr(contar(hs, func(h models.ParticipanteHabilidad) bool { return enRango(h.Fecha, f.Desde, f.Hasta) }))
		series["habilidades"] = []models.SerieGrafica{{Nombre: "Habilidades desarrolladas", Puntos: porMes(func(m Mes) int {
			return contar(hs, func(h models.ParticipanteHabilidad) bool { return !h.Fecha.Before(m.Inicio) && !h.Fecha.After(m.Corte) })
		})}}
	}

	if d.Postulaciones != nil {
		ins := filtrar(*d.Postulaciones, func(p models.Postulacion) bool { return p.Estado == "seleccionado" })
		valores["insercionesLaborales"] = ptr(contar(ins, func(p models.Postulacion) bool { return enRango(p.FechaEstado, f.Desde, f.Hasta) }))
		series["inserciones"] = []models.SerieGrafica{{Nombre: "Inserciones laborales", Puntos: porMes(func(m Mes) int {
			return contar(ins, func(p models.Postulacion) bool {
				return !p.FechaEstado.Before(m.Inicio) && !p.FechaEstado.After(m.Corte)
			})
		})}}
	}

	// El área formativa no aplica a Promoción Humana.
	if d.Promocion != nil {
		pr := *d.Promocion
		fichas := func(ref time.Time) int {
			return contar(pr.Fichas, func(fi models.FichaPromocion) bool { return FichaActivaEn(fi, ref) })
		}
		valores["fichasActivas"] = ptr(fichas(corte))
		valores["participantesConSeguimiento"] = ptr(ParticipantesConSeguimiento(pr.Hitos, f.Desde, f.Hasta))
		series["promocion"] = []models.SerieGrafica{
			{Nombre: "Fichas activas", Puntos: porMes(func(m Mes) int { return fichas(m.Corte) })},
			// Cada mes: participantes con algún hito en ese mes, sin
			// salirse del período elegido.
			{Nombre: "Con seguimiento", Puntos: porMes(func(m Mes) int { return ParticipantesConSeguimiento(pr.Hitos, m.Inicio, m.Corte) })},
		}
	}

	visible := func(modulo string) bool { return f.Modulo == ModuloTodos || f.Modulo == modulo }

	out := models.Resumen{Indicadores: []models.Indicador{}, Graficas: []models.Grafica{}}
	for _, m := range metaIndicadores {
		if !visible(m.modulo) {
			continue
		}
		v := valores[m.id]
		out.Indicadores = append(out.Indicadores, models.Indicador{
			ID: m.id, Titulo: m.titulo, Descripcion: m.descripcion, Modulo: m.modulo, Fuente: m.fuente,
			Valor: v, Disponible: v != nil,
		})
	}
	for _, g := range []models.Grafica{
		{ID: "participantes", Titulo: "Participantes activos", Tipo: "line", Modulo: ModuloUsuarios, Indicadores: []string{"participantesActivos"}},
		{ID: "cursos", Titulo: "Cursos en desarrollo y finalizados", Tipo: "bar", Modulo: ModuloEducacion, Indicadores: []string{"cursosEnDesarrollo", "cursosFinalizados"}},
		{ID: "habilidades", Titulo: "Habilidades desarrolladas por mes", Tipo: "line", Modulo: ModuloEducacion, Indicadores: []string{"habilidadesDesarrolladas"}},
		{ID: "inserciones", Titulo: "Inserciones laborales por mes", Tipo: "line", Modulo: ModuloInsercion, Indicadores: []string{"insercionesLaborales"}},
		{ID: "promocion", Titulo: "Fichas activas y participantes con seguimiento", Tipo: "bar", Modulo: ModuloPromocion, Indicadores: []string{"fichasActivas", "participantesConSeguimiento"}},
	} {
		if !visible(g.Modulo) {
			continue
		}
		g.Series = series[g.ID]
		out.Graficas = append(out.Graficas, g)
	}
	return out
}
