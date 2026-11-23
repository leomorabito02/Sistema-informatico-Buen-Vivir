package services

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"buenvivir/dashboard/internal/models"
)

// Mismos casos que tenía el frontend (dashboard.service.spec.ts) cuando el
// cálculo vivía allá: las reglas no cambiaron, solo de lugar.

// d es el mediodía del día, en la hora del centro.
func d(s string) time.Time {
	t, err := time.ParseInLocation("2006-01-02", s, Zona)
	if err != nil {
		panic(err)
	}
	return t.Add(12 * time.Hour)
}

func dia(s string) time.Time { return d(s).Add(-12 * time.Hour) }

func pt(t time.Time) *time.Time { return &t }

func filtros(desde, hasta, modulo, area string) Filtros {
	return Filtros{Desde: dia(desde), Hasta: dia(hasta), Modulo: modulo, AreaID: area}
}

var semestre = filtros("2026-01-01", "2026-06-30", ModuloTodos, "")

var participantesPrueba = []models.Participante{
	{ID: "p1", Rol: "participante", Estado: "activo", FechaAlta: d("2025-12-01")},
	{ID: "p2", Rol: "participante", Estado: "activo", FechaAlta: d("2026-03-10")},
	// Baja lógica en abril: cuenta hasta marzo.
	{ID: "p3", Rol: "participante", Estado: "baja", FechaAlta: d("2025-11-01"), FechaBaja: pt(d("2026-04-15"))},
	// Alta posterior al período: no cuenta.
	{ID: "p4", Rol: "participante", Estado: "activo", FechaAlta: d("2026-08-01")},
}

func curso(id, area string, fechas []string, estado, creacion string) models.Curso {
	c := models.Curso{ID: id, Titulo: id, AreaID: area, Estado: estado, FechaCreacion: d(creacion), Actividades: []models.Actividad{}}
	for _, f := range fechas {
		c.Actividades = append(c.Actividades, models.Actividad{ID: id + "-" + f, Fecha: d(f)})
	}
	return c
}

var cursosPrueba = []models.Curso{
	curso("c1", "huerta", []string{"2026-02-01", "2026-03-01"}, "activo", "2026-01-05"), // finaliza en marzo
	curso("c2", "huerta", []string{"2026-02-01", "2026-09-01"}, "activo", "2026-01-05"), // pendiente al 30/06
	curso("c3", "oficios", []string{"2026-04-01"}, "activo", "2026-01-05"),              // finaliza en abril
	curso("c4", "oficios", []string{"2026-02-01"}, "borrador", "2026-01-05"),            // borrador: no cuenta
}

func hab(id, area, fecha, estado string) models.ParticipanteHabilidad {
	return models.ParticipanteHabilidad{ID: id, ParticipanteID: "p1", CursoID: "c1", AreaID: area, Origen: "actividad", Estado: estado, Fecha: d(fecha)}
}

var habilidadesPrueba = []models.ParticipanteHabilidad{
	hab("h1", "huerta", "2026-02-10", "aceptada"),
	hab("h2", "huerta", "2026-02-20", "aceptada"),
	hab("h3", "oficios", "2026-05-02", "aceptada"),
	hab("h4", "huerta", "2026-03-01", "pendiente"),
	hab("h5", "huerta", "2026-03-01", "rechazada"),
	hab("h6", "huerta", "2025-12-01", "aceptada"), // fuera del período
}

func post(id, estado, fecha string) models.Postulacion {
	return models.Postulacion{ID: id, ParticipanteID: "p1", Estado: estado, FechaEstado: d(fecha)}
}

var postulacionesPrueba = []models.Postulacion{
	post("s1", "seleccionado", "2026-01-20"),
	post("s2", "seleccionado", "2026-06-29"),
	post("s3", "rechazado", "2026-02-01"),
	post("s4", "postulado", "2026-03-01"),
	post("s5", "seleccionado", "2026-07-15"), // fuera del período
}

// f2 se dio de baja en abril; f3 todavía no existía.
var promocionPrueba = models.Promocion{
	Fichas: []models.FichaPromocion{
		{ID: "f1", ParticipanteID: "p1", Estado: "activa", FechaAlta: d("2026-01-10")},
		{ID: "f2", ParticipanteID: "p2", Estado: "baja", FechaAlta: d("2026-02-01"), FechaBaja: pt(d("2026-04-15"))},
		{ID: "f3", ParticipanteID: "p3", Estado: "activa", FechaAlta: d("2026-08-01")},
	},
	Hitos: []models.HitoPromocion{
		{ParticipanteID: "p1", Fecha: "2026-01-20"},
		{ParticipanteID: "p1", Fecha: "2026-03-05"},
		{ParticipanteID: "p2", Fecha: "2026-03-31"},
		{ParticipanteID: "p3", Fecha: "2026-07-01"}, // fuera del período
		{ParticipanteID: "p2", Fecha: "2025-12-31"}, // fuera del período
	},
}

func todos() Datos {
	ps, cs, hs, po, pr := participantesPrueba, cursosPrueba, habilidadesPrueba, postulacionesPrueba, promocionPrueba
	return Datos{Participantes: &ps, Cursos: &cs, Habilidades: &hs, Postulaciones: &po, Promocion: &pr, Equipo: &models.Equipo{Empresas: 5, Docentes: 6, Voluntarios: 7}}
}

func valor(t *testing.T, r models.Resumen, id string) *int {
	t.Helper()
	for _, i := range r.Indicadores {
		if i.ID == id {
			return i.Valor
		}
	}
	t.Fatalf("no está el indicador %s", id)
	return nil
}

func esperar(t *testing.T, r models.Resumen, id string, want int) {
	t.Helper()
	if v := valor(t, r, id); v == nil || *v != want {
		t.Errorf("%s: esperaba %d, obtuve %v", id, want, v)
	}
}

func grafica(t *testing.T, r models.Resumen, id string) models.Grafica {
	t.Helper()
	for _, g := range r.Graficas {
		if g.ID == id {
			return g
		}
	}
	t.Fatalf("no está la gráfica %s", id)
	return models.Grafica{}
}

func valores(s models.SerieGrafica) []int {
	out := make([]int, len(s.Puntos))
	for i, p := range s.Puntos {
		out[i] = p.Valor
	}
	return out
}

func TestEstadoCurso(t *testing.T) {
	c := curso("x", "a", []string{"2026-02-01", "2026-03-01"}, "activo", "2026-01-05")
	if got := EstadoCurso(c, d("2026-02-15")); got != cursoEnDesarrollo {
		t.Errorf("con actividades pendientes: %s", got)
	}
	if got := EstadoCurso(c, d("2026-03-02")); got != cursoFinalizado {
		t.Errorf("con todas las actividades ocurridas: %s", got)
	}
	if got := EstadoCurso(curso("b", "a", []string{"2026-02-01"}, "borrador", "2026-01-05"), d("2026-06-01")); got != cursoFuera {
		t.Errorf("borrador: %s", got)
	}
	if got := EstadoCurso(curso("f", "a", []string{"2026-09-01"}, "activo", "2026-08-01"), d("2026-06-01")); got != cursoFuera {
		t.Errorf("todavía no existía: %s", got)
	}
	if got := EstadoCurso(curso("v", "a", nil, "activo", "2026-01-05"), d("2026-06-01")); got != cursoEnDesarrollo {
		t.Errorf("activo sin actividades: %s", got)
	}
}

func TestParticipanteActivoEn(t *testing.T) {
	p3 := participantesPrueba[2]
	if !ParticipanteActivoEn(p3, d("2026-03-31")) || ParticipanteActivoEn(p3, d("2026-04-30")) {
		t.Error("la baja lógica cuenta desde su fecha")
	}
	sinFecha := p3
	sinFecha.FechaBaja = nil
	if ParticipanteActivoEn(sinFecha, d("2026-01-01")) {
		t.Error("una baja sin fecha no cuenta")
	}
	otroRol := participantesPrueba[0]
	otroRol.Rol = "voluntario"
	if ParticipanteActivoEn(otroRol, d("2026-06-01")) {
		t.Error("solo cuentan participantes")
	}
}

func TestFichaActivaEn(t *testing.T) {
	f1, f2, f3 := promocionPrueba.Fichas[0], promocionPrueba.Fichas[1], promocionPrueba.Fichas[2]
	casos := []struct {
		f    models.FichaPromocion
		ref  string
		want bool
	}{
		{f1, "2026-01-09", false}, {f1, "2026-06-30", true},
		{f2, "2026-03-31", true}, {f2, "2026-04-30", false},
		{f3, "2026-06-30", false},
	}
	for _, c := range casos {
		if got := FichaActivaEn(c.f, d(c.ref)); got != c.want {
			t.Errorf("%s al %s: esperaba %v", c.f.ID, c.ref, c.want)
		}
	}
	sinFecha := f2
	sinFecha.FechaBaja = nil
	if FichaActivaEn(sinFecha, d("2026-03-01")) {
		t.Error("una baja sin fecha no cuenta")
	}
}

func TestParticipantesConSeguimiento(t *testing.T) {
	if got := ParticipantesConSeguimiento(promocionPrueba.Hitos, dia("2026-01-01"), dia("2026-06-30")); got != 2 {
		t.Errorf("personas distintas en el período: %d", got)
	}
	// Un hito del último día cuenta aunque el corte sea de noche.
	noche := dia("2026-06-30").Add(23 * time.Hour)
	if got := ParticipantesConSeguimiento([]models.HitoPromocion{{ParticipanteID: "x", Fecha: "2026-06-30"}}, dia("2026-01-01"), noche); got != 1 {
		t.Errorf("último día: %d", got)
	}
	if got := ParticipantesConSeguimiento([]models.HitoPromocion{{ParticipanteID: "x", Fecha: "2026-01-01"}}, dia("2026-01-02"), dia("2026-06-30")); got != 0 {
		t.Errorf("antes del período: %d", got)
	}
	if got := ParticipantesConSeguimiento(nil, dia("2026-01-01"), dia("2026-06-30")); got != 0 {
		t.Errorf("sin hitos: %d", got)
	}
}

func TestMesesDelPeriodo(t *testing.T) {
	meses := MesesDelPeriodo(dia("2026-01-15"), dia("2026-03-10"))
	var claves []string
	for _, m := range meses {
		claves = append(claves, m.Clave)
	}
	if strings.Join(claves, ",") != "2026-01,2026-02,2026-03" {
		t.Fatalf("meses: %v", claves)
	}
	if meses[0].Inicio.Day() != 15 || meses[2].Corte.Day() != 10 {
		t.Errorf("el primer mes empieza en desde y el último corta en hasta: %v / %v", meses[0].Inicio, meses[2].Corte)
	}
	if len(MesesDelPeriodo(dia("2026-06-01"), dia("2026-01-01"))) != 0 {
		t.Error("período invertido: sin meses")
	}
	if MesesEntre(dia("2016-01-01"), dia("2026-12-31")) != 132 || MesesEntre(dia("2026-01-31"), dia("2026-02-01")) != 2 {
		t.Error("MesesEntre")
	}
}

func TestCalcularResumen_Indicadores(t *testing.T) {
	r := CalcularResumen(todos(), semestre)
	var ids []string
	for _, i := range r.Indicadores {
		ids = append(ids, i.ID)
		if !i.Disponible {
			t.Errorf("%s debería estar disponible", i.ID)
		}
	}
	want := []string{"participantesActivos", "empresasActivas", "docentesActivos", "voluntariosActivos", "cursosEnDesarrollo", "cursosFinalizados", "habilidadesDesarrolladas", "insercionesLaborales", "fichasActivas", "participantesConSeguimiento"}
	if !reflect.DeepEqual(ids, want) {
		t.Fatalf("orden de los indicadores: %v", ids)
	}
	esperar(t, r, "participantesActivos", 2)        // p1, p2
	esperar(t, r, "empresasActivas", 5)
	esperar(t, r, "docentesActivos", 6)
	esperar(t, r, "voluntariosActivos", 7)
	esperar(t, r, "cursosEnDesarrollo", 1)          // c2
	esperar(t, r, "cursosFinalizados", 2)           // c1, c3
	esperar(t, r, "habilidadesDesarrolladas", 3)    // h1, h2, h3
	esperar(t, r, "insercionesLaborales", 2)        // s1, s2
	esperar(t, r, "fichasActivas", 1)               // f1
	esperar(t, r, "participantesConSeguimiento", 2) // p1, p2
}

func TestCalcularResumen_Graficas(t *testing.T) {
	r := CalcularResumen(todos(), semestre)
	var ids []string
	for _, g := range r.Graficas {
		ids = append(ids, g.ID)
	}
	if strings.Join(ids, ",") != "participantes,cursos,habilidades,inserciones,promocion" {
		t.Fatalf("gráficas: %v", ids)
	}

	promo := grafica(t, r, "promocion")
	if got := valores(promo.Series[0]); !reflect.DeepEqual(got, []int{1, 2, 2, 1, 1, 1}) {
		t.Errorf("fichas activas por mes: %v", got)
	}
	if got := valores(promo.Series[1]); !reflect.DeepEqual(got, []int{1, 0, 2, 0, 0, 0}) {
		t.Errorf("con seguimiento por mes: %v", got)
	}

	cursos := grafica(t, r, "cursos")
	if cursos.Tipo != "bar" || cursos.Series[0].Nombre != "En desarrollo" || cursos.Series[1].Nombre != "Finalizados" {
		t.Errorf("gráfica de cursos: %+v", cursos)
	}
	if len(cursos.Series[0].Puntos) != 6 || cursos.Series[0].Puntos[1].Valor != 3 || cursos.Series[1].Puntos[2].Valor != 1 {
		t.Errorf("febrero: 3 en desarrollo; marzo: c1 finalizado: %+v", cursos.Series)
	}

	if got := valores(grafica(t, r, "participantes").Series[0]); !reflect.DeepEqual(got, []int{2, 2, 3, 2, 2, 2}) {
		t.Errorf("participantes por mes: %v", got)
	}
	if got := valores(grafica(t, r, "habilidades").Series[0]); !reflect.DeepEqual(got, []int{0, 2, 0, 0, 1, 0}) {
		t.Errorf("habilidades por mes: %v", got)
	}
	if got := valores(grafica(t, r, "inserciones").Series[0]); !reflect.DeepEqual(got, []int{1, 0, 0, 0, 0, 1}) {
		t.Errorf("inserciones por mes: %v", got)
	}
	if grafica(t, r, "inserciones").Series[0].Puntos[0].Mes != "2026-01" {
		t.Error("los meses van como AAAA-MM")
	}
}

func TestCalcularResumen_Periodo(t *testing.T) {
	r := CalcularResumen(todos(), filtros("2026-01-01", "2026-02-28", ModuloTodos, ""))
	esperar(t, r, "cursosFinalizados", 0)
	esperar(t, r, "cursosEnDesarrollo", 3)
	esperar(t, r, "habilidadesDesarrolladas", 2)
	esperar(t, r, "insercionesLaborales", 1)
	if n := len(r.Graficas[0].Series[0].Puntos); n != 2 {
		t.Errorf("dos meses, dos puntos: %d", n)
	}
}

func TestCalcularResumen_Area(t *testing.T) {
	r := CalcularResumen(todos(), filtros("2026-01-01", "2026-06-30", ModuloTodos, "oficios"))
	esperar(t, r, "cursosEnDesarrollo", 0)
	esperar(t, r, "cursosFinalizados", 1)
	esperar(t, r, "habilidadesDesarrolladas", 1)
	// El área no aplica a participantes ni inserciones.
	esperar(t, r, "participantesActivos", 2)
	esperar(t, r, "insercionesLaborales", 2)
}

func TestCalcularResumen_Modulo(t *testing.T) {
	ids := func(r models.Resumen) (ind, gra string) {
		var a, b []string
		for _, i := range r.Indicadores {
			a = append(a, i.ID)
		}
		for _, g := range r.Graficas {
			b = append(b, g.ID)
		}
		return strings.Join(a, ","), strings.Join(b, ",")
	}
	casos := map[string][2]string{
		ModuloEducacion: {"cursosEnDesarrollo,cursosFinalizados,habilidadesDesarrolladas", "cursos,habilidades"},
		ModuloInsercion: {"insercionesLaborales", "inserciones"},
		ModuloPromocion: {"fichasActivas,participantesConSeguimiento", "promocion"},
		ModuloUsuarios:  {"participantesActivos,empresasActivas,docentesActivos,voluntariosActivos", "participantes"},
	}
	for modulo, want := range casos {
		ind, gra := ids(CalcularResumen(todos(), filtros("2026-01-01", "2026-06-30", modulo, "")))
		if ind != want[0] || gra != want[1] {
			t.Errorf("%s: %s / %s", modulo, ind, gra)
		}
	}
	if !ModuloValido(ModuloTodos) || ModuloValido("otro") {
		t.Error("ModuloValido")
	}
}

func TestCalcularResumen_FuenteCaida(t *testing.T) {
	datos := todos()
	datos.Cursos = nil
	r := CalcularResumen(datos, semestre)
	var nd []string
	for _, i := range r.Indicadores {
		if !i.Disponible {
			nd = append(nd, i.ID)
		}
	}
	if strings.Join(nd, ",") != "cursosEnDesarrollo,cursosFinalizados" {
		t.Errorf("no disponibles: %v", nd)
	}
	if valor(t, r, "cursosEnDesarrollo") != nil {
		t.Error("sin datos, el valor es null")
	}
	esperar(t, r, "participantesActivos", 2)
	if grafica(t, r, "cursos").Series != nil || grafica(t, r, "participantes").Series == nil {
		t.Error("solo la gráfica de la fuente caída queda sin series")
	}

	// En JSON, valor y series salen como null (el frontend muestra "No disponible").
	b, _ := json.Marshal(r)
	if !strings.Contains(string(b), `"id":"cursosEnDesarrollo","titulo":"Cursos en desarrollo","descripcion":"Cursos publicados con actividades pendientes.","modulo":"educacion","fuente":"educacion","valor":null,"disponible":false`) {
		t.Errorf("JSON del indicador no disponible: %s", b)
	}
	if !strings.Contains(string(b), `"series":null`) {
		t.Errorf("JSON de la gráfica sin datos: %s", b)
	}
}

func TestCalcularResumen_HoraDelCentro(t *testing.T) {
	// Un alta del 31/03 a las 22 h (hora del centro) es de marzo, aunque
	// en UTC ya sea 1/04.
	alta := time.Date(2026, 3, 31, 22, 0, 0, 0, Zona)
	ps := []models.Participante{{ID: "x", Rol: "participante", Estado: "activo", FechaAlta: alta}}
	r := CalcularResumen(Datos{Participantes: &ps}, filtros("2026-03-01", "2026-04-30", ModuloUsuarios, ""))
	if got := valores(r.Graficas[0].Series[0]); !reflect.DeepEqual(got, []int{1, 1}) {
		t.Errorf("el alta cuenta en marzo: %v", got)
	}
}
