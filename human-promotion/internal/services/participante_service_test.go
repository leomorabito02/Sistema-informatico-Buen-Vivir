package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"buenvivir/human-promotion/internal/models"
	"buenvivir/human-promotion/internal/repositories"
)

// authPorUsuario resuelve user_id -> participante como lo haría Auth.
type authPorUsuario struct {
	fakeAuth
	participantes map[string]string // user_id -> participante_id
	err           error
}

func (a authPorUsuario) GetParticipanteByUserID(_ context.Context, userID string) (*models.ParticipanteInfo, error) {
	if a.err != nil {
		return nil, a.err
	}
	id, ok := a.participantes[userID]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	return &models.ParticipanteInfo{ParticipanteID: id, UserID: userID}, nil
}

type hitosFalsos struct {
	repositories.HitoRepository
	porFicha map[string]map[string]models.HitoDocument
}

func (h hitosFalsos) ListByFicha(_ context.Context, fichaID string) (map[string]models.HitoDocument, error) {
	return h.porFicha[fichaID], nil
}

// cifradoFalso "descifra" sacando el prefijo enc:.
type cifradoFalso struct{}

func (cifradoFalso) Encrypt(s string) (string, error) { return "enc:" + s, nil }
func (cifradoFalso) Decrypt(s string) (string, error) {
	if !strings.HasPrefix(s, "enc:") {
		return "", errors.New("no cifrado")
	}
	return strings.TrimPrefix(s, "enc:"), nil
}

func nuevoParticipanteService(linea lineaFalsa) ParticipanteService {
	fichas := fichasFalsas{
		porID:           map[string]models.FichaDocument{"f1": {ParticipanteID: "p1"}},
		porParticipante: map[string]string{"p1": "f1"},
	}
	hitos := hitosFalsos{porFicha: map[string]map[string]models.HitoDocument{"f1": {
		"h2": {FichaID: "f1", Fecha: "2026-09-10", Descripcion: "enc:Empezó el taller de cocina", VisibleParticipante: true},
		"h1": {FichaID: "f1", Fecha: "2026-09-01", Descripcion: "enc:Primera entrevista", VisibleParticipante: true},
		"h3": {FichaID: "f1", Fecha: "2026-09-05", Descripcion: "enc:CONFIDENCIAL", VisibleParticipante: false},
	}}}
	auth := authPorUsuario{participantes: map[string]string{"u1": "p1", "u2": "p2"}}
	return NewParticipanteService(auth, fichas, hitos, cifradoFalso{}, NewActividadService(linea, fichas))
}

func TestLineaDeTiempoPropia_HitosVisiblesYActividad(t *testing.T) {
	svc := nuevoParticipanteService(lineaFalsa{entradas: map[string][]repositories.EntradaLineaTiempo{
		"p1": {{ID: "i1", Tipo: repositories.LineaTiempoInscripcion, Descripcion: "Se inscribió en un curso", Fecha: "2026-09-02", CursoID: "c1"}},
	}})

	linea, err := svc.LineaDeTiempoPropia(context.Background(), "u1")
	if err != nil {
		t.Fatal(err)
	}
	if len(linea.Hitos) != 2 || linea.Hitos[0].ID != "h1" || linea.Hitos[1].Descripcion != "Empezó el taller de cocina" {
		t.Errorf("hitos visibles, descifrados y en orden: %+v", linea.Hitos)
	}
	for _, h := range linea.Hitos {
		if h.Descripcion == "CONFIDENCIAL" {
			t.Fatal("un hito no visible llegó al participante")
		}
	}
	if len(linea.Actividad) != 1 || linea.Actividad[0].CursoID != "c1" {
		t.Errorf("actividad: %+v", linea.Actividad)
	}
}

func TestLineaDeTiempoPropia_SinFichaConActividad(t *testing.T) {
	svc := nuevoParticipanteService(lineaFalsa{entradas: map[string][]repositories.EntradaLineaTiempo{
		"p2": {{ID: "a1", Tipo: repositories.LineaTiempoAsistenciaGeneral, Fecha: "2026-09-02"}},
	}})
	linea, err := svc.LineaDeTiempoPropia(context.Background(), "u2")
	if err != nil {
		t.Fatal(err)
	}
	if linea.Hitos == nil || len(linea.Hitos) != 0 {
		t.Errorf("sin ficha los hitos vienen vacíos (no nil): %+v", linea.Hitos)
	}
	if len(linea.Actividad) != 1 {
		t.Errorf("la actividad se muestra aunque no tenga ficha: %+v", linea.Actividad)
	}
}

func TestLineaDeTiempoPropia_NoEsParticipante(t *testing.T) {
	linea, err := nuevoParticipanteService(lineaFalsa{}).LineaDeTiempoPropia(context.Background(), "desconocido")
	if err != nil || len(linea.Hitos) != 0 || len(linea.Actividad) != 0 || linea.Actividad == nil {
		t.Fatalf("esperaba listas vacías, obtuve %+v, %v", linea, err)
	}
}

func TestLineaDeTiempoPropia_ActividadCaida(t *testing.T) {
	svc := nuevoParticipanteService(lineaFalsa{err: errors.New("firestore caído")})
	linea, err := svc.LineaDeTiempoPropia(context.Background(), "u1")
	if err != nil {
		t.Fatalf("la línea tiene que salir igual con los hitos: %v", err)
	}
	if len(linea.Hitos) != 2 || len(linea.Actividad) != 0 {
		t.Errorf("esperaba solo hitos: %+v", linea)
	}
}

func TestLineaDeTiempoPropia_AuthCaido(t *testing.T) {
	svc := NewParticipanteService(authPorUsuario{err: errors.New("auth caído")}, fichasFalsas{}, hitosFalsos{}, cifradoFalso{}, nil)
	if _, err := svc.LineaDeTiempoPropia(context.Background(), "u1"); err == nil {
		t.Fatal("esperaba el error de Auth")
	}
}

// authBusqueda registra lo que se le pide a Auth en la búsqueda.
type authBusqueda struct {
	fakeAuth
	query  *string
	limite *int
}

func (a authBusqueda) SearchParticipantes(_ context.Context, q string, limit int) ([]models.ParticipanteInfo, error) {
	*a.query, *a.limite = q, limit
	return []models.ParticipanteInfo{
		{ParticipanteID: "p1", Nombre: "Ana", Apellido: "Pérez"},
		{ParticipanteID: "p9", Nombre: "Luis", Apellido: "Gómez"},
	}, nil
}

// Sin texto, Buscar devuelve el listado completo con el tope del listado; con 1 letra
// sigue siendo error de validación.
func TestBuscar_SinTextoListaTodosConFicha(t *testing.T) {
	var q string
	var limite int
	fichas := fichasFalsas{porParticipante: map[string]string{"p1": "f1"}}
	svc := NewParticipanteService(authBusqueda{query: &q, limite: &limite}, fichas, hitosFalsos{}, cifradoFalso{}, nil)

	res, err := svc.Buscar(context.Background(), "   ")
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if q != "" || limite != maxListadoParticipantes {
		t.Fatalf("a Auth le llegó q=%q limit=%d", q, limite)
	}
	if len(res) != 2 || res[0].FichaID == nil || *res[0].FichaID != "f1" || res[1].FichaID != nil {
		t.Fatalf("resultado inesperado: %+v", res)
	}

	if _, err := svc.Buscar(context.Background(), "ana"); err != nil || limite != maxResultadosBusqueda {
		t.Fatalf("búsqueda con texto: err=%v limit=%d", err, limite)
	}

	var ve *ValidationError
	if _, err := svc.Buscar(context.Background(), "a"); !errors.As(err, &ve) {
		t.Fatalf("1 letra debería ser error de validación, fue %v", err)
	}
}

func (f fichasFalsas) FindIDsByParticipantes(_ context.Context, ids []string) (map[string]string, error) {
	out := map[string]string{}
	for _, id := range ids {
		if fid, ok := f.porParticipante[id]; ok {
			out[id] = fid
		}
	}
	return out, nil
}
