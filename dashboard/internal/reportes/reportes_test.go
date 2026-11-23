package reportes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
)

var pdfValido = []byte("%PDF-1.7\nreporte de prueba")

// ── Dobles de prueba ───────────────────────────────────────────────────

type repoMem struct {
	datos   map[string]Reporte
	sig     int
	errCrea error
}

func newRepoMem() *repoMem { return &repoMem{datos: map[string]Reporte{}} }

func (r *repoMem) NuevoID() string { r.sig++; return fmt.Sprintf("rep%d", r.sig) }
func (r *repoMem) Crear(_ context.Context, rep Reporte) error {
	if r.errCrea != nil {
		return r.errCrea
	}
	r.datos[rep.ID] = rep
	return nil
}
func (r *repoMem) Obtener(_ context.Context, id string) (Reporte, error) {
	rep, ok := r.datos[id]
	if !ok {
		return Reporte{}, ErrNoEncontrado
	}
	return rep, nil
}
func (r *repoMem) Listar(_ context.Context, limite int) ([]Reporte, error) {
	out := []Reporte{}
	for _, v := range r.datos {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].FechaGeneracion.After(out[j].FechaGeneracion) })
	if len(out) > limite {
		out = out[:limite]
	}
	return out, nil
}

type storageMem struct {
	objetos  map[string][]byte
	errSubir error
	borrados []string
}

func newStorageMem() *storageMem { return &storageMem{objetos: map[string][]byte{}} }

func (s *storageMem) Subir(_ context.Context, objeto, _ string, data []byte) error {
	if s.errSubir != nil {
		return s.errSubir
	}
	s.objetos[objeto] = data
	return nil
}
func (s *storageMem) Abrir(_ context.Context, objeto string) (io.ReadCloser, error) {
	d, ok := s.objetos[objeto]
	if !ok {
		return nil, ErrNoEncontrado
	}
	return io.NopCloser(bytes.NewReader(d)), nil
}
func (s *storageMem) Borrar(_ context.Context, objeto string) error {
	s.borrados = append(s.borrados, objeto)
	delete(s.objetos, objeto)
	return nil
}

func metaValida() Metadatos {
	area := "Huerta"
	return Metadatos{PeriodoDesde: "2026-01-01", PeriodoHasta: "2026-10-20", FiltrosAplicados: Filtros{Modulo: "todos", AreaNombre: &area}, GeneradoPorNombre: " Ana Pérez "}
}

// ── Servicio ───────────────────────────────────────────────────────────

func TestRegistrar_GuardaDatosYPDF(t *testing.T) {
	repo, st := newRepoMem(), newStorageMem()
	svc := NewService(repo, st, "ceibuenvivir-bucket")
	svc.ahora = func() time.Time { return time.Date(2026, 10, 20, 15, 0, 0, 0, time.FixedZone("ART", -3*3600)) }

	rep, err := svc.Registrar(context.Background(), metaValida(), pdfValido, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	if rep.ID != "rep1" || rep.TipoReporte != "dashboard" || rep.GeneradoPor != "admin-1" || rep.GeneradoPorNombre != "Ana Pérez" {
		t.Errorf("reporte: %+v", rep)
	}
	if rep.URLArchivo != "gs://ceibuenvivir-bucket/reportes/rep1.pdf" || rep.FechaGeneracion.Location() != time.UTC {
		t.Errorf("ubicación o fecha: %+v", rep)
	}
	if !bytes.Equal(st.objetos["reportes/rep1.pdf"], pdfValido) {
		t.Error("el PDF no se subió al bucket, en la carpeta reportes/")
	}
	if _, ok := repo.datos["rep1"]; !ok {
		t.Error("los datos no se guardaron")
	}
}

func TestRegistrar_Validaciones(t *testing.T) {
	largo := strings.Repeat("a", 121)
	casos := []struct {
		nombre  string
		cambiar func(*Metadatos, *[]byte, *string)
		mensaje string
	}{
		{"sin usuario", func(_ *Metadatos, _ *[]byte, u *string) { *u = " " }, "No se pudo identificar a quien exporta el reporte"},
		{"no es PDF", func(_ *Metadatos, p *[]byte, _ *string) { *p = []byte("<html>") }, "El archivo tiene que ser un PDF"},
		{"vacío", func(_ *Metadatos, p *[]byte, _ *string) { *p = nil }, "El archivo tiene que ser un PDF"},
		{"muy grande", func(_ *Metadatos, p *[]byte, _ *string) { *p = append([]byte("%PDF-"), make([]byte, MaxPDFBytes)...) }, "El PDF no puede superar los 10 MB"},
		{"fecha mal", func(m *Metadatos, _ *[]byte, _ *string) { m.PeriodoDesde = "01/01/2026" }, "El período tiene que tener fechas AAAA-MM-DD"},
		{"período al revés", func(m *Metadatos, _ *[]byte, _ *string) { m.PeriodoHasta = "2025-12-31" }, "La fecha de fin del período no puede ser anterior a la de inicio"},
		{"módulo raro", func(m *Metadatos, _ *[]byte, _ *string) { m.FiltrosAplicados.Modulo = "otro" }, "Módulo de filtro desconocido"},
		{"nombre largo", func(m *Metadatos, _ *[]byte, _ *string) { m.GeneradoPorNombre = largo }, "El nombre de quien exporta es demasiado largo"},
		{"área larga", func(m *Metadatos, _ *[]byte, _ *string) { m.FiltrosAplicados.AreaNombre = &largo }, "El filtro de área es demasiado largo"},
	}
	for _, tc := range casos {
		t.Run(tc.nombre, func(t *testing.T) {
			meta, pdf, usuario := metaValida(), pdfValido, "admin-1"
			tc.cambiar(&meta, &pdf, &usuario)
			st := newStorageMem()
			_, err := NewService(newRepoMem(), st, "b").Registrar(context.Background(), meta, pdf, usuario)
			var v *ErrValidacion
			if !errors.As(err, &v) || v.Mensaje != tc.mensaje {
				t.Fatalf("esperaba %q, obtuve %v", tc.mensaje, err)
			}
			if len(st.objetos) != 0 {
				t.Error("con datos inválidos no se tiene que subir nada")
			}
		})
	}
}

func TestRegistrar_SiFallaGuardarBorraElPDF(t *testing.T) {
	repo, st := newRepoMem(), newStorageMem()
	repo.errCrea = errors.New("firestore caído")
	if _, err := NewService(repo, st, "b").Registrar(context.Background(), metaValida(), pdfValido, "admin-1"); err == nil {
		t.Fatal("esperaba error")
	}
	if len(st.objetos) != 0 || len(st.borrados) != 1 {
		t.Fatalf("el PDF quedó suelto en el bucket: %v / borrados %v", st.objetos, st.borrados)
	}

	st2 := newStorageMem()
	st2.errSubir = errors.New("bucket caído")
	repo2 := newRepoMem()
	if _, err := NewService(repo2, st2, "b").Registrar(context.Background(), metaValida(), pdfValido, "admin-1"); err == nil {
		t.Fatal("esperaba error al subir")
	}
	if len(repo2.datos) != 0 {
		t.Fatal("sin PDF no se tienen que guardar los datos")
	}
}

func TestAbrirPDF(t *testing.T) {
	repo, st := newRepoMem(), newStorageMem()
	svc := NewService(repo, st, "b")
	rep, _ := svc.Registrar(context.Background(), metaValida(), pdfValido, "admin-1")

	_, rc, err := svc.AbrirPDF(context.Background(), rep.ID)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(rc)
	if !bytes.Equal(b, pdfValido) {
		t.Error("PDF distinto")
	}
	for _, id := range []string{"no-existe", "../otro", ""} {
		if _, _, err := svc.AbrirPDF(context.Background(), id); !errors.Is(err, ErrNoEncontrado) {
			t.Errorf("%q: esperaba ErrNoEncontrado, obtuve %v", id, err)
		}
	}
}

// ── HTTP ───────────────────────────────────────────────────────────────

func multipartReporte(t *testing.T, meta any, pdf []byte) (*bytes.Buffer, string) {
	t.Helper()
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	if meta != nil {
		m, _ := json.Marshal(meta)
		_ = w.WriteField("metadatos", string(m))
	}
	if pdf != nil {
		f, _ := w.CreateFormFile("archivo", "reporte.pdf")
		_, _ = f.Write(pdf)
	}
	_ = w.Close()
	return &b, w.FormDataContentType()
}

func muxHandlers(h *Handlers) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /reportes", h.Registrar)
	mux.HandleFunc("GET /reportes", h.Listar)
	mux.HandleFunc("GET /reportes/{id}/pdf", h.Descargar)
	return mux
}

func TestHandlers_RegistrarListarYDescargar(t *testing.T) {
	mux := muxHandlers(NewHandlers(NewService(newRepoMem(), newStorageMem(), "b")))

	body, ct := multipartReporte(t, metaValida(), pdfValido)
	req := httptest.NewRequest(http.MethodPost, "/reportes", body)
	req.Header.Set("Content-Type", ct)
	req.Header.Set("X-User-ID", "admin-1")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("registrar: %d %s", rec.Code, rec.Body.String())
	}
	var rep Reporte
	_ = json.Unmarshal(rec.Body.Bytes(), &rep)

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/reportes", nil))
	var lista []Reporte
	if err := json.Unmarshal(rec.Body.Bytes(), &lista); err != nil || len(lista) != 1 || lista[0].ID != rep.ID {
		t.Fatalf("listar: %s", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/reportes/"+rep.ID+"/pdf", nil))
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "application/pdf" ||
		!strings.Contains(rec.Header().Get("Content-Disposition"), "attachment") || !bytes.Equal(rec.Body.Bytes(), pdfValido) {
		t.Fatalf("descargar: %d %v", rec.Code, rec.Header())
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/reportes/no-existe/pdf", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("inexistente: %d", rec.Code)
	}
}

func TestHandlers_PedidosInvalidos(t *testing.T) {
	mux := muxHandlers(NewHandlers(NewService(newRepoMem(), newStorageMem(), "b")))
	casos := map[string]func() (*bytes.Buffer, string){
		"sin metadatos":   func() (*bytes.Buffer, string) { return multipartReporte(t, nil, pdfValido) },
		"sin PDF":         func() (*bytes.Buffer, string) { return multipartReporte(t, metaValida(), nil) },
		"metadatos rotos": func() (*bytes.Buffer, string) { return multipartReporte(t, "texto", pdfValido) },
		"no es multipart": func() (*bytes.Buffer, string) { return bytes.NewBufferString("{}"), "application/json" },
		"archivo no PDF":  func() (*bytes.Buffer, string) { return multipartReporte(t, metaValida(), []byte("hola")) },
	}
	for nombre, armar := range casos {
		body, ct := armar()
		req := httptest.NewRequest(http.MethodPost, "/reportes", body)
		req.Header.Set("Content-Type", ct)
		req.Header.Set("X-User-ID", "admin-1")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: esperaba 400, obtuve %d %s", nombre, rec.Code, rec.Body.String())
		}
	}
}

func TestHandlers_ErrorInternoNoExponeDetalles(t *testing.T) {
	repo := newRepoMem()
	repo.errCrea = errors.New("rpc error: firestore.googleapis.com")
	mux := muxHandlers(NewHandlers(NewService(repo, newStorageMem(), "b")))
	body, ct := multipartReporte(t, metaValida(), pdfValido)
	req := httptest.NewRequest(http.MethodPost, "/reportes", body)
	req.Header.Set("Content-Type", ct)
	req.Header.Set("X-User-ID", "admin-1")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError || strings.Contains(rec.Body.String(), "firestore") {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
}

// ── Firestore (emulador) ───────────────────────────────────────────────

func TestFirestoreRepository(t *testing.T) {
	host := os.Getenv("FIRESTORE_EMULATOR_HOST")
	if host == "" {
		t.Skip("FIRESTORE_EMULATOR_HOST no seteado: se omite la prueba contra el emulador")
	}
	if c, err := net.DialTimeout("tcp", host, 2*time.Second); err != nil {
		t.Skipf("emulador no accesible: %v", err)
	} else {
		_ = c.Close()
	}
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "dashboard-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()
	docs, _ := client.Collection(Collection).Documents(ctx).GetAll()
	for _, d := range docs {
		_, _ = d.Ref.Delete(ctx)
	}

	repo := NewFirestoreRepository(client)
	area := "Huerta"
	for i, h := range []int{10, 12, 11} {
		r := Reporte{ID: repo.NuevoID(), TipoReporte: "dashboard", PeriodoDesde: "2026-01-01", PeriodoHasta: "2026-10-20",
			FiltrosAplicados: Filtros{Modulo: "educacion", AreaNombre: &area}, GeneradoPor: fmt.Sprintf("u%d", i),
			FechaGeneracion: time.Date(2026, 10, 20, h, 0, 0, 0, time.UTC), URLArchivo: "gs://b/reportes/x.pdf"}
		if err := repo.Crear(ctx, r); err != nil {
			t.Fatal(err)
		}
	}
	lista, err := repo.Listar(ctx, 2)
	if err != nil || len(lista) != 2 || lista[0].GeneradoPor != "u1" || lista[1].GeneradoPor != "u2" {
		t.Fatalf("listar ordenado y con límite: %+v %v", lista, err)
	}
	got, err := repo.Obtener(ctx, lista[0].ID)
	if err != nil || got.FiltrosAplicados.Modulo != "educacion" || got.FiltrosAplicados.AreaNombre == nil || *got.FiltrosAplicados.AreaNombre != "Huerta" || got.FiltrosAplicados.AreaID != nil {
		t.Fatalf("obtener: %+v %v", got, err)
	}
	if _, err := repo.Obtener(ctx, "no-existe"); !errors.Is(err, ErrNoEncontrado) {
		t.Fatalf("inexistente: %v", err)
	}
}

func TestNuevoSinBucket(t *testing.T) {
	svc, cerrar, err := Nuevo(context.Background(), "")
	if svc != nil || cerrar != nil || err != nil {
		t.Fatalf("sin bucket no se arma nada: %v %v", svc, err)
	}
}

func TestResolveFirestoreDatabase(t *testing.T) {
	for _, tc := range []struct {
		name string
		val  string
		want string
	}{
		{"empty", "", ""},
		{"default", "(default)", ""},
		{"null", "null", ""},
		{"custom", "buenvivir-prod", "buenvivir-prod"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("FIRESTORE_DATABASE", tc.val)
			if got := resolveFirestoreDatabase(); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
