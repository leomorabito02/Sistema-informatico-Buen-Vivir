package services

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// memContents es un TeacherContentRepository en memoria.
type memContents struct {
	items     map[string]models.ContenidoDocente
	descargas []models.DescargaBiblioteca
	next      int
	createErr error
}

func newMemContents() *memContents { return &memContents{items: map[string]models.ContenidoDocente{}} }

func (m *memContents) NewID() string { m.next++; return "cont-" + string(rune('0'+m.next)) }
func (m *memContents) Create(_ context.Context, c *models.ContenidoDocente) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.items[c.ID] = *c
	return nil
}
func (m *memContents) Get(_ context.Context, id string) (*models.ContenidoDocente, error) {
	c, ok := m.items[id]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	return &c, nil
}
func (m *memContents) filtrar(f func(models.ContenidoDocente) bool) []models.ContenidoDocente {
	out := []models.ContenidoDocente{}
	for _, c := range m.items {
		if f(c) {
			out = append(out, c)
		}
	}
	return out
}
func (m *memContents) ListByDocente(_ context.Context, d string) ([]models.ContenidoDocente, error) {
	return m.filtrar(func(c models.ContenidoDocente) bool { return c.DocenteID == d }), nil
}
func (m *memContents) ListByEstado(_ context.Context, e string) ([]models.ContenidoDocente, error) {
	return m.filtrar(func(c models.ContenidoDocente) bool { return e == "" || c.EstadoModeracion == e }), nil
}
func (m *memContents) Update(_ context.Context, c *models.ContenidoDocente) error {
	m.items[c.ID] = *c
	return nil
}
func (m *memContents) Moderate(_ context.Context, id, estado, motivo, adminID string) (*models.ContenidoDocente, error) {
	c, ok := m.items[id]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	c.EstadoModeracion, c.MotivoRechazo, c.ModeradoPor = estado, motivo, adminID
	m.items[id] = c
	return &c, nil
}
func (m *memContents) Delete(_ context.Context, id string) error { delete(m.items, id); return nil }
func (m *memContents) CreateDescarga(_ context.Context, d *models.DescargaBiblioteca) error {
	m.descargas = append(m.descargas, *d)
	return nil
}
func (m *memContents) ListDescargas(context.Context) ([]models.DescargaBiblioteca, error) {
	return m.descargas, nil
}

// memStorage es un FileStorage en memoria.
type memStorage struct{ files map[string][]byte }

func newMemStorage() *memStorage { return &memStorage{files: map[string][]byte{}} }
func (s *memStorage) Upload(_ context.Context, path, _ string, data []byte) error {
	s.files[path] = data
	return nil
}
func (s *memStorage) Open(_ context.Context, path string) (io.ReadCloser, error) {
	d, ok := s.files[path]
	if !ok {
		return nil, errors.New("no existe")
	}
	return io.NopCloser(bytes.NewReader(d)), nil
}
func (s *memStorage) DeletePrefix(_ context.Context, prefix string) error {
	for k := range s.files {
		if strings.HasPrefix(k, prefix) {
			delete(s.files, k)
		}
	}
	return nil
}

func pdf(nombre string) *models.ArchivoSubido {
	return &models.ArchivoSubido{Nombre: nombre, Tipo: "application/pdf", Data: []byte("%PDF-1.4 contenido")}
}

func nuevoContentSvc() (*teacherContentService, *memContents, *memStorage) {
	repo, st := newMemContents(), newMemStorage()
	s := &teacherContentService{repo: repo, storage: st, now: func() time.Time { return time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC) }}
	return s, repo, st
}

func TestTeacherContent_UploadQuedaPendiente(t *testing.T) {
	s, repo, st := nuevoContentSvc()
	c, err := s.Upload(context.Background(), "usr-doc", models.ContenidoInput{Titulo: " Guía de huerta ", Descripcion: "<b>Paso a paso</b>"}, pdf("guía huerta.pdf"))
	require.NoError(t, err)
	assert.Equal(t, models.ModeracionPendiente, c.EstadoModeracion)
	assert.Equal(t, "Guía de huerta", c.Titulo)
	assert.Equal(t, "Paso a paso", c.Descripcion, "sin HTML")
	assert.Equal(t, "usr-doc", c.DocenteID)
	assert.True(t, strings.HasPrefix(c.ArchivoRuta, "contenidos/"+c.ID+"/"), c.ArchivoRuta)
	assert.NotContains(t, c.ArchivoRuta, " ", "nombre de archivo saneado")
	assert.Len(t, st.files, 1)
	assert.Len(t, repo.items, 1)
}

func TestTeacherContent_UploadValidaciones(t *testing.T) {
	ok := models.ContenidoInput{Titulo: "Guía"}
	casos := map[string]struct {
		in      models.ContenidoInput
		archivo *models.ArchivoSubido
	}{
		"sin archivo":       {ok, nil},
		"tipo no válido":    {ok, &models.ArchivoSubido{Nombre: "x.exe", Tipo: "application/x-msdownload", Data: []byte("MZ")}},
		"archivo grande":    {ok, &models.ArchivoSubido{Nombre: "g.pdf", Tipo: "application/pdf", Data: make([]byte, maxArchivoContenido+1)}},
		"título corto":      {models.ContenidoInput{Titulo: "ab"}, pdf("a.pdf")},
		"título largo":      {models.ContenidoInput{Titulo: strings.Repeat("a", 121)}, pdf("a.pdf")},
		"descripción larga": {models.ContenidoInput{Titulo: "Guía", Descripcion: strings.Repeat("a", 1001)}, pdf("a.pdf")},
	}
	for nombre, c := range casos {
		t.Run(nombre, func(t *testing.T) {
			s, repo, st := nuevoContentSvc()
			_, err := s.Upload(context.Background(), "usr-doc", c.in, c.archivo)
			assert.True(t, IsValidationError(err), "vino %v", err)
			assert.Empty(t, repo.items)
			assert.Empty(t, st.files)
		})
	}
}

func TestTeacherContent_SinBucketResponde503(t *testing.T) {
	s := &teacherContentService{repo: newMemContents(), now: time.Now}
	_, err := s.Upload(context.Background(), "usr-doc", models.ContenidoInput{Titulo: "Guía"}, pdf("a.pdf"))
	assert.True(t, IsUnavailableError(err))
}

func TestTeacherContent_SiFallaGuardarSeBorraElArchivo(t *testing.T) {
	s, repo, st := nuevoContentSvc()
	repo.createErr = errors.New("firestore caído")
	_, err := s.Upload(context.Background(), "usr-doc", models.ContenidoInput{Titulo: "Guía"}, pdf("a.pdf"))
	assert.Error(t, err)
	assert.Empty(t, st.files)
}

func TestTeacherContent_EditarSoloDuenoYPendiente(t *testing.T) {
	s, repo, st := nuevoContentSvc()
	c, err := s.Upload(context.Background(), "usr-doc", models.ContenidoInput{Titulo: "Guía"}, pdf("a.pdf"))
	require.NoError(t, err)
	rutaVieja := c.ArchivoRuta

	_, err = s.Update(context.Background(), "otro", c.ID, models.ContenidoInput{Titulo: "Ajena"}, nil)
	assert.True(t, IsForbiddenError(err))

	s.now = func() time.Time { return time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC) }
	ed, err := s.Update(context.Background(), "usr-doc", c.ID, models.ContenidoInput{Titulo: "Guía nueva"}, pdf("b.pdf"))
	require.NoError(t, err)
	assert.Equal(t, "Guía nueva", ed.Titulo)
	assert.Equal(t, "b.pdf", ed.ArchivoNombre)
	assert.NotContains(t, st.files, rutaVieja, "el archivo viejo se borra")
	assert.Contains(t, st.files, ed.ArchivoRuta)

	_, err = s.Moderate(context.Background(), "usr-admin", c.ID, models.ModeracionAprobado, "")
	require.NoError(t, err)
	_, err = s.Update(context.Background(), "usr-doc", c.ID, models.ContenidoInput{Titulo: "Otra"}, nil)
	assert.True(t, IsValidationError(err), "aprobado ya no se edita")
	assert.Equal(t, "Guía nueva", repo.items[c.ID].Titulo)
}

func TestTeacherContent_Eliminar(t *testing.T) {
	s, repo, st := nuevoContentSvc()
	c, _ := s.Upload(context.Background(), "usr-doc", models.ContenidoInput{Titulo: "Guía"}, pdf("a.pdf"))
	_, _ = s.Moderate(context.Background(), "usr-admin", c.ID, models.ModeracionAprobado, "")

	assert.True(t, IsValidationError(s.Delete(context.Background(), "usr-doc", "role_teacher", c.ID)), "el dueño no borra lo aprobado")
	assert.True(t, IsForbiddenError(s.Delete(context.Background(), "otro", "role_teacher", c.ID)))
	require.NoError(t, s.Delete(context.Background(), "usr-admin", "role_admin", c.ID), "el admin borra cualquiera")
	assert.Empty(t, repo.items)
	assert.Empty(t, st.files)

	p, _ := s.Upload(context.Background(), "usr-doc", models.ContenidoInput{Titulo: "Pendiente"}, pdf("p.pdf"))
	require.NoError(t, s.Delete(context.Background(), "usr-doc", "role_teacher", p.ID), "el dueño borra lo pendiente")
}

func TestTeacherContent_Moderar(t *testing.T) {
	s, _, _ := nuevoContentSvc()
	c, _ := s.Upload(context.Background(), "usr-doc", models.ContenidoInput{Titulo: "Guía"}, pdf("a.pdf"))

	_, err := s.Moderate(context.Background(), "usr-admin", c.ID, "publicado", "")
	assert.True(t, IsValidationError(err))
	_, err = s.Moderate(context.Background(), "usr-admin", c.ID, models.ModeracionRechazado, strings.Repeat("a", 501))
	assert.True(t, IsValidationError(err))

	r, err := s.Moderate(context.Background(), "usr-admin", c.ID, models.ModeracionRechazado, " Falta la fuente <i>x</i> ")
	require.NoError(t, err)
	assert.Equal(t, "Falta la fuente x", r.MotivoRechazo)

	a, err := s.Moderate(context.Background(), "usr-admin", c.ID, models.ModeracionAprobado, "no se guarda")
	require.NoError(t, err)
	assert.Empty(t, a.MotivoRechazo, "aprobar limpia el motivo")

	_, err = s.Moderate(context.Background(), "usr-admin", "no-existe", models.ModeracionAprobado, "")
	assert.ErrorIs(t, err, repositories.ErrNotFound)

	_, err = s.List(context.Background(), "cualquiera")
	assert.True(t, IsValidationError(err))
}

func TestTeacherContent_BibliotecaPublica(t *testing.T) {
	s, repo, _ := nuevoContentSvc()
	pend, _ := s.Upload(context.Background(), "usr-doc", models.ContenidoInput{Titulo: "Pendiente"}, pdf("p.pdf"))
	apr, _ := s.Upload(context.Background(), "usr-doc", models.ContenidoInput{Titulo: "Aprobado"}, pdf("a.pdf"))
	_, _ = s.Moderate(context.Background(), "usr-admin", apr.ID, models.ModeracionAprobado, "")

	pub, err := s.ListPublic(context.Background())
	require.NoError(t, err)
	require.Len(t, pub, 1, "solo lo aprobado")
	assert.Equal(t, "Aprobado", pub[0].Titulo)

	datos := models.DescargaInput{Nombre: "Ana", Apellido: "Gómez", Email: "ana@gmail.com", Telefono: "(3547) 15-123 456"}
	_, _, err = s.PublicDownload(context.Background(), pend.ID, datos)
	assert.ErrorIs(t, err, repositories.ErrNotFound, "lo pendiente no existe para el público")

	c, rc, err := s.PublicDownload(context.Background(), apr.ID, datos)
	require.NoError(t, err)
	body, _ := io.ReadAll(rc)
	_ = rc.Close()
	assert.Equal(t, "a.pdf", c.ArchivoNombre)
	assert.Equal(t, "%PDF-1.4 contenido", string(body))
	require.Len(t, repo.descargas, 1)
	assert.Equal(t, "354715123456", repo.descargas[0].Telefono, "teléfono normalizado")
	assert.Equal(t, apr.ID, repo.descargas[0].ContenidoID)
}

func TestTeacherContent_DescargaValidaDatos(t *testing.T) {
	ok := models.DescargaInput{Nombre: "Ana", Apellido: "Gómez", Email: "ana@gmail.com", Telefono: "3547123456"}
	casos := map[string]func(*models.DescargaInput){
		"sin nombre":       func(d *models.DescargaInput) { d.Nombre = "" },
		"sin apellido":     func(d *models.DescargaInput) { d.Apellido = " " },
		"email inválido":   func(d *models.DescargaInput) { d.Email = "ana" },
		"email con nombre": func(d *models.DescargaInput) { d.Email = "Ana <ana@gmail.com>" },
		"teléfono corto":   func(d *models.DescargaInput) { d.Telefono = "1234" },
		"teléfono letras":  func(d *models.DescargaInput) { d.Telefono = "3547abc456" },
	}
	for nombre, cambiar := range casos {
		t.Run(nombre, func(t *testing.T) {
			s, repo, _ := nuevoContentSvc()
			in := ok
			cambiar(&in)
			_, _, err := s.PublicDownload(context.Background(), "cualquiera", in)
			assert.True(t, IsValidationError(err), "vino %v", err)
			assert.Empty(t, repo.descargas)
		})
	}
}

func TestNormalizarTelefono(t *testing.T) {
	for entrada, quiero := range map[string]string{
		"3547 123456":        "3547123456",
		"+54 9 3547 12-3456": "+5493547123456",
		"(0351) 456.7890":    "03514567890",
	} {
		got, ok := NormalizarTelefono(entrada)
		assert.True(t, ok, entrada)
		assert.Equal(t, quiero, got)
	}
	for _, mal := range []string{"", "1234567", "12345678901234567", "35+47123456"} {
		_, ok := NormalizarTelefono(mal)
		assert.False(t, ok, mal)
	}
}

func TestTeacherContent_ArchivoConSesion(t *testing.T) {
	s, _, _ := nuevoContentSvc()
	c, _ := s.Upload(context.Background(), "usr-doc", models.ContenidoInput{Titulo: "Guía"}, pdf("a.pdf"))

	_, _, err := s.File(context.Background(), "otro", "role_teacher", c.ID)
	assert.True(t, IsForbiddenError(err), "pendiente: no lo ve otro docente")

	for _, quien := range [][2]string{{"usr-doc", "role_teacher"}, {"usr-admin", "role_admin"}} {
		_, rc, err := s.File(context.Background(), quien[0], quien[1], c.ID)
		require.NoError(t, err, quien[1])
		_ = rc.Close()
	}

	_, _ = s.Moderate(context.Background(), "usr-admin", c.ID, models.ModeracionAprobado, "")
	_, rc, err := s.File(context.Background(), "otro", "role_teacher", c.ID)
	require.NoError(t, err, "aprobado: lo consulta cualquier usuario con sesión (HU-22)")
	_ = rc.Close()
}
