package services

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	"buenvivir/campaigns/internal/models"
	"buenvivir/campaigns/internal/repositories"
)

// fakeRepo es un CampaignRepository en memoria, con el mismo contrato que
// el de Firestore (ErrNotFound, orden por fecha descendente).
type fakeRepo struct {
	datos map[string]models.Campaign
	sig   int
	err   error
}

func newFakeRepo() *fakeRepo { return &fakeRepo{datos: map[string]models.Campaign{}} }

func (f *fakeRepo) Create(_ context.Context, c models.Campaign) (models.Campaign, error) {
	if f.err != nil {
		return models.Campaign{}, f.err
	}
	f.sig++
	c.ID = fmt.Sprintf("cmp%d", f.sig)
	f.datos[c.ID] = c
	return c, nil
}

func (f *fakeRepo) Get(_ context.Context, id string) (models.Campaign, error) {
	c, ok := f.datos[id]
	if !ok {
		return models.Campaign{}, repositories.ErrNotFound
	}
	return c, nil
}

func (f *fakeRepo) Update(_ context.Context, id string, in models.CampaignInput) (models.Campaign, error) {
	c, ok := f.datos[id]
	if !ok {
		return models.Campaign{}, repositories.ErrNotFound
	}
	c.Titulo, c.Texto, c.ImagenURL, c.Fecha = in.Titulo, in.Texto, in.ImagenURL, in.Fecha
	f.datos[id] = c
	return c, nil
}

func (f *fakeRepo) SetEstado(_ context.Context, id, estado string) (models.Campaign, error) {
	c, ok := f.datos[id]
	if !ok {
		return models.Campaign{}, repositories.ErrNotFound
	}
	c.Estado = estado
	f.datos[id] = c
	return c, nil
}

func (f *fakeRepo) List(_ context.Context, estado string) ([]models.Campaign, error) {
	if f.err != nil {
		return nil, f.err
	}
	var out []models.Campaign
	for _, c := range f.datos {
		if estado == "" || c.Estado == estado {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Fecha > out[j].Fecha })
	return out, nil
}

// Delete simula la baja lógica: la campaña deja de existir para Get y List.
func (f *fakeRepo) Delete(_ context.Context, id, _ string) error {
	if _, ok := f.datos[id]; !ok {
		return repositories.ErrNotFound
	}
	delete(f.datos, id)
	return nil
}

func entradaValida() models.CampaignInput {
	return models.CampaignInput{
		Titulo:    "Jornada de reciclaje",
		Texto:     "Traé tus botellas al centro.",
		ImagenURL: "https://example.org/reciclaje.jpg",
		Fecha:     "2026-11-05",
	}
}

func TestCreate_QuedaEnBorradorConAutor(t *testing.T) {
	svc := NewCampaignService(newFakeRepo())
	c, err := svc.Create(context.Background(), entradaValida(), "usuario-1")
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if c.ID == "" || c.Estado != models.EstadoBorrador || c.CreadoPor != "usuario-1" {
		t.Fatalf("campaña creada mal: %+v", c)
	}
}

func TestCreate_LimpiaTextos(t *testing.T) {
	svc := NewCampaignService(newFakeRepo())
	in := entradaValida()
	in.Titulo = "  <b>Jornada</b> de reciclaje\x07  "
	in.Texto = "Línea 1\r\n<script>alert(1)</script>Línea 2"
	c, err := svc.Create(context.Background(), in, "usuario-1")
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if c.Titulo != "Jornada de reciclaje" {
		t.Errorf("título sin limpiar: %q", c.Titulo)
	}
	if strings.Contains(c.Texto, "<") || !strings.Contains(c.Texto, "Línea 1\n") {
		t.Errorf("texto sin limpiar: %q", c.Texto)
	}
}

func TestCreate_Validaciones(t *testing.T) {
	casos := []struct {
		nombre  string
		cambiar func(*models.CampaignInput)
		mensaje string
	}{
		{"faltan todos", func(in *models.CampaignInput) { *in = models.CampaignInput{} }, "Completar los campos obligatorios: título, texto, imagen, fecha"},
		{"falta la imagen", func(in *models.CampaignInput) { in.ImagenURL = "  " }, "Completar los campos obligatorios: imagen"},
		{"título solo con HTML", func(in *models.CampaignInput) { in.Titulo = "<br>" }, "Completar los campos obligatorios: título"},
		{"título largo", func(in *models.CampaignInput) { in.Titulo = strings.Repeat("a", 151) }, "El título no puede superar los 150 caracteres"},
		{"texto largo", func(in *models.CampaignInput) { in.Texto = strings.Repeat("ñ", 5001) }, "El texto no puede superar los 5000 caracteres"},
		{"imagen javascript", func(in *models.CampaignInput) { in.ImagenURL = "javascript:alert(1)" }, "La imagen debe ser una dirección web válida (http o https)"},
		{"imagen sin host", func(in *models.CampaignInput) { in.ImagenURL = "https://" }, "La imagen debe ser una dirección web válida (http o https)"},
		{"imagen data svg", func(in *models.CampaignInput) { in.ImagenURL = "data:image/svg+xml;base64,PHN2Zz4=" }, "La imagen debe ser una dirección web válida (http o https)"},
		{"imagen ftp", func(in *models.CampaignInput) { in.ImagenURL = "ftp://example.org/a.jpg" }, "La imagen debe ser una dirección web válida (http o https)"},
		{"fecha con barras", func(in *models.CampaignInput) { in.Fecha = "05/11/2026" }, "La fecha debe tener el formato AAAA-MM-DD"},
		{"fecha imposible", func(in *models.CampaignInput) { in.Fecha = "2026-02-30" }, "La fecha debe tener el formato AAAA-MM-DD"},
		{"año fuera de rango", func(in *models.CampaignInput) { in.Fecha = "1800-01-01" }, "La fecha debe tener el formato AAAA-MM-DD"},
	}
	for _, tc := range casos {
		t.Run(tc.nombre, func(t *testing.T) {
			in := entradaValida()
			tc.cambiar(&in)
			_, err := NewCampaignService(newFakeRepo()).Create(context.Background(), in, "usuario-1")
			if !IsValidationError(err) || err.Error() != tc.mensaje {
				t.Fatalf("esperaba validación %q, obtuve %v", tc.mensaje, err)
			}
		})
	}
}

func TestCreate_SinUsuario(t *testing.T) {
	_, err := NewCampaignService(newFakeRepo()).Create(context.Background(), entradaValida(), " ")
	if !IsValidationError(err) {
		t.Fatalf("esperaba error de validación, obtuve %v", err)
	}
}

func TestUpdate_EditaYConservaEstado(t *testing.T) {
	repo := newFakeRepo()
	svc := NewCampaignService(repo)
	c, _ := svc.Create(context.Background(), entradaValida(), "usuario-1")
	if _, err := svc.Publish(context.Background(), c.ID); err != nil {
		t.Fatal(err)
	}
	in := entradaValida()
	in.Titulo = "Jornada de reciclaje (nueva fecha)"
	editada, err := svc.Update(context.Background(), c.ID, in)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if editada.Titulo != in.Titulo || editada.Estado != models.EstadoPublicada || editada.CreadoPor != "usuario-1" {
		t.Fatalf("edición incorrecta: %+v", editada)
	}
}

func TestUpdate_ErroresDeEntrada(t *testing.T) {
	svc := NewCampaignService(newFakeRepo())
	if _, err := svc.Update(context.Background(), "../otra", entradaValida()); !errors.Is(err, repositories.ErrNotFound) {
		t.Errorf("id inválido: esperaba ErrNotFound, obtuve %v", err)
	}
	if _, err := svc.Update(context.Background(), "noexiste", entradaValida()); !errors.Is(err, repositories.ErrNotFound) {
		t.Errorf("id inexistente: esperaba ErrNotFound, obtuve %v", err)
	}
	if _, err := svc.Update(context.Background(), "cmp1", models.CampaignInput{}); !IsValidationError(err) {
		t.Errorf("campos vacíos: esperaba validación, obtuve %v", err)
	}
}

func TestPublish_PasaAPublicadaYEsIdempotente(t *testing.T) {
	svc := NewCampaignService(newFakeRepo())
	c, _ := svc.Create(context.Background(), entradaValida(), "usuario-1")
	for i := 0; i < 2; i++ {
		p, err := svc.Publish(context.Background(), c.ID)
		if err != nil || p.Estado != models.EstadoPublicada {
			t.Fatalf("publicación %d: %+v, %v", i+1, p, err)
		}
	}
	if _, err := svc.Publish(context.Background(), "noexiste"); !errors.Is(err, repositories.ErrNotFound) {
		t.Fatalf("esperaba ErrNotFound, obtuve %v", err)
	}
}

func TestListPublished_SoloPublicadasYBusquedaPorTitulo(t *testing.T) {
	svc := NewCampaignService(newFakeRepo())
	ctx := context.Background()
	crear := func(titulo, fecha string, publicar bool) {
		in := entradaValida()
		in.Titulo, in.Fecha = titulo, fecha
		c, err := svc.Create(ctx, in, "usuario-1")
		if err != nil {
			t.Fatal(err)
		}
		if publicar {
			if _, err := svc.Publish(ctx, c.ID); err != nil {
				t.Fatal(err)
			}
		}
	}
	crear("Campaña del Árbol", "2026-09-01", true)
	crear("Jornada de reciclaje", "2026-10-01", true)
	crear("Campaña en borrador", "2026-12-01", false)

	todas, err := svc.ListPublished(ctx, "")
	if err != nil || len(todas) != 2 || todas[0].Titulo != "Jornada de reciclaje" {
		t.Fatalf("publicadas: %+v, %v", todas, err)
	}
	// Sin tildes ni mayúsculas encuentra igual (HU-51, coincidencia parcial).
	for _, q := range []string{"campana", "ARBOL", "  árbol "} {
		got, err := svc.ListPublished(ctx, q)
		if err != nil || len(got) != 1 || got[0].Titulo != "Campaña del Árbol" {
			t.Errorf("búsqueda %q: %+v, %v", q, got, err)
		}
	}
	if got, _ := svc.ListPublished(ctx, "inexistente"); len(got) != 0 {
		t.Errorf("búsqueda sin coincidencias devolvió %d", len(got))
	}
	if _, err := svc.ListPublished(ctx, strings.Repeat("a", 101)); !IsValidationError(err) {
		t.Errorf("búsqueda larga: esperaba validación, obtuve %v", err)
	}
}

func TestList_GestionVeTodoYFiltraPorEstado(t *testing.T) {
	svc := NewCampaignService(newFakeRepo())
	ctx := context.Background()
	c, _ := svc.Create(ctx, entradaValida(), "usuario-1")
	_, _ = svc.Create(ctx, entradaValida(), "usuario-2")
	_, _ = svc.Publish(ctx, c.ID)

	for estado, esperado := range map[string]int{"": 2, "borrador": 1, "publicada": 1} {
		got, err := svc.List(ctx, "", estado)
		if err != nil || len(got) != esperado {
			t.Errorf("estado %q: %d campañas, %v", estado, len(got), err)
		}
	}
	if _, err := svc.List(ctx, "", "archivada"); !IsValidationError(err) {
		t.Errorf("estado inválido: esperaba validación, obtuve %v", err)
	}
}

func TestErroresDelRepositorioSePropagan(t *testing.T) {
	repo := newFakeRepo()
	repo.err = errors.New("firestore caído")
	svc := NewCampaignService(repo)
	if _, err := svc.Create(context.Background(), entradaValida(), "usuario-1"); err == nil || IsValidationError(err) {
		t.Errorf("create: esperaba el error del repositorio, obtuve %v", err)
	}
	if _, err := svc.ListPublished(context.Background(), ""); err == nil {
		t.Error("list: esperaba el error del repositorio")
	}
}

type mockCampanasProducer struct {
	createdID    string
	authorID     string
	notifyCalled bool
	llamadas     int
}

func (m *mockCampanasProducer) Enabled() bool {
	return true
}

func (m *mockCampanasProducer) CampanaCreada(_ context.Context, campanaID, autorID string) error {
	m.createdID = campanaID
	m.authorID = autorID
	m.llamadas++
	return nil
}

func (m *mockCampanasProducer) Notify() {
	m.notifyCalled = true
}

// El aviso de campaña nueva sale una sola vez, al publicar: crear el
// borrador no avisa y publicar de nuevo tampoco (Problema 3 de la
// auditoría: antes avisaba dos veces, la primera de un borrador).
func TestCampanaCreada_SoloAlPublicarYUnaVez(t *testing.T) {
	prod := &mockCampanasProducer{}
	svc := NewCampaignService(
		newFakeRepo(),
		WithCampanasProducer(prod),
	)
	ctx := context.Background()
	c, err := svc.Create(ctx, entradaValida(), "usuario-admin")
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if prod.llamadas != 0 {
		t.Fatalf("crear el borrador no debería avisar: %d eventos", prod.llamadas)
	}
	if _, err := svc.Publish(ctx, c.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Publish(ctx, c.ID); err != nil {
		t.Fatal(err)
	}
	if prod.llamadas != 1 {
		t.Fatalf("publicar debería avisar exactamente una vez: %d eventos", prod.llamadas)
	}
	if prod.createdID != c.ID || prod.authorID != "usuario-admin" {
		t.Errorf("evento campana.creada incorrecto: id=%s, author=%s", prod.createdID, prod.authorID)
	}
	if !prod.notifyCalled {
		t.Errorf("esperaba llamada a Notify()")
	}
}

// ── Baja lógica (Problema 5 de la auditoría 2026-11-11) ──

func TestDelete_SacaLaCampanaDeGestionYPublico(t *testing.T) {
	ctx := context.Background()
	svc := NewCampaignService(newFakeRepo())
	c, _ := svc.Create(ctx, entradaValida(), "usuario-1")
	if _, err := svc.Publish(ctx, c.ID); err != nil {
		t.Fatal(err)
	}
	if err := svc.Delete(ctx, c.ID, "admin-1"); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if lista, _ := svc.List(ctx, "", ""); len(lista) != 0 {
		t.Fatalf("la gestión no debería listar la campaña dada de baja: %v", lista)
	}
	if pub, _ := svc.ListPublished(ctx, ""); len(pub) != 0 {
		t.Fatalf("la landing no debería mostrar la campaña dada de baja: %v", pub)
	}
	if _, err := svc.Update(ctx, c.ID, entradaValida()); !errors.Is(err, repositories.ErrNotFound) {
		t.Fatalf("editar una campaña dada de baja debería dar ErrNotFound, dio %v", err)
	}
	if _, err := svc.Publish(ctx, c.ID); !errors.Is(err, repositories.ErrNotFound) {
		t.Fatalf("publicar una campaña dada de baja debería dar ErrNotFound, dio %v", err)
	}
}

func TestDelete_Validaciones(t *testing.T) {
	ctx := context.Background()
	svc := NewCampaignService(newFakeRepo())
	if err := svc.Delete(ctx, "../otra", "admin-1"); !errors.Is(err, repositories.ErrNotFound) {
		t.Fatalf("id inválido debería dar ErrNotFound, dio %v", err)
	}
	if err := svc.Delete(ctx, "noexiste", "admin-1"); !errors.Is(err, repositories.ErrNotFound) {
		t.Fatalf("id inexistente debería dar ErrNotFound, dio %v", err)
	}
	c, _ := svc.Create(ctx, entradaValida(), "usuario-1")
	if err := svc.Delete(ctx, c.ID, " "); !IsValidationError(err) {
		t.Fatalf("sin usuario debería ser error de validación, dio %v", err)
	}
	if err := svc.Delete(ctx, c.ID, "admin-1"); err != nil {
		t.Fatal(err)
	}
	if err := svc.Delete(ctx, c.ID, "admin-1"); !errors.Is(err, repositories.ErrNotFound) {
		t.Fatalf("dar de baja dos veces debería dar ErrNotFound, dio %v", err)
	}
}
