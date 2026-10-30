package controllers_test

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"buenvivir/education/internal/controllers"
	"buenvivir/education/internal/middlewares"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"
)

type fakeContentService struct {
	err       error
	got       struct{ actor, role, id, estado, motivo string }
	gotIn     models.ContenidoInput
	gotFile   *models.ArchivoSubido
	gotDesc   models.DescargaInput
	contenido models.ContenidoDocente
}

func (f *fakeContentService) Upload(_ context.Context, d string, in models.ContenidoInput, a *models.ArchivoSubido) (*models.ContenidoDocente, error) {
	f.got.actor, f.gotIn, f.gotFile = d, in, a
	return &f.contenido, f.err
}
func (f *fakeContentService) ListMine(_ context.Context, d string) ([]models.ContenidoDocente, error) {
	f.got.actor = d
	return []models.ContenidoDocente{f.contenido}, f.err
}
func (f *fakeContentService) Update(_ context.Context, d, id string, in models.ContenidoInput, a *models.ArchivoSubido) (*models.ContenidoDocente, error) {
	f.got.actor, f.got.id, f.gotIn, f.gotFile = d, id, in, a
	return &f.contenido, f.err
}
func (f *fakeContentService) Delete(_ context.Context, a, r, id string) error {
	f.got.actor, f.got.role, f.got.id = a, r, id
	return f.err
}
func (f *fakeContentService) List(_ context.Context, e string) ([]models.ContenidoDocente, error) {
	f.got.estado = e
	return []models.ContenidoDocente{}, f.err
}
func (f *fakeContentService) Moderate(_ context.Context, a, id, e, m string) (*models.ContenidoDocente, error) {
	f.got.actor, f.got.id, f.got.estado, f.got.motivo = a, id, e, m
	return &f.contenido, f.err
}
func (f *fakeContentService) ListDownloads(context.Context) ([]models.DescargaBiblioteca, error) {
	return []models.DescargaBiblioteca{}, f.err
}
func (f *fakeContentService) ListPublic(context.Context) ([]models.ContenidoPublico, error) {
	return []models.ContenidoPublico{f.contenido.Publico()}, f.err
}
func (f *fakeContentService) PublicDownload(_ context.Context, id string, in models.DescargaInput) (*models.ContenidoDocente, io.ReadCloser, error) {
	f.got.id, f.gotDesc = id, in
	if f.err != nil {
		return nil, nil, f.err
	}
	return &f.contenido, io.NopCloser(strings.NewReader("contenido")), nil
}
func (f *fakeContentService) File(_ context.Context, a, r, id string) (*models.ContenidoDocente, io.ReadCloser, error) {
	f.got.actor, f.got.role, f.got.id = a, r, id
	if f.err != nil {
		return nil, nil, f.err
	}
	return &f.contenido, io.NopCloser(strings.NewReader("contenido")), nil
}

func conActor(r *http.Request, user, role, id string) *http.Request {
	ctx := context.WithValue(r.Context(), middlewares.UserIDKey, user)
	ctx = context.WithValue(ctx, middlewares.RoleIDKey, role)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return r.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))
}

func multipartMaterial(t *testing.T, conArchivo bool) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	require.NoError(t, mw.WriteField("titulo", "Guía"))
	require.NoError(t, mw.WriteField("descripcion", "Desc"))
	if conArchivo {
		fw, err := mw.CreateFormFile("archivo", "guia.pdf")
		require.NoError(t, err)
		_, _ = fw.Write([]byte("%PDF"))
	}
	require.NoError(t, mw.Close())
	return &buf, mw.FormDataContentType()
}

func TestTeacherContentController_UploadYUpdate(t *testing.T) {
	svc := &fakeContentService{contenido: models.ContenidoDocente{ID: "c1", ArchivoRuta: "secreta"}}
	ctrl := controllers.NewTeacherContentController(svc)

	body, ct := multipartMaterial(t, true)
	req := httptest.NewRequest("POST", "/api/contenidos-docentes", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	ctrl.Upload(w, conActor(req, "usr-doc", "role_teacher", ""))
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.NotContains(t, w.Body.String(), "secreta", "la ruta del archivo no se expone")
	assert.Equal(t, "usr-doc", svc.got.actor)
	assert.Equal(t, "Guía", svc.gotIn.Titulo)
	require.NotNil(t, svc.gotFile)
	assert.Equal(t, "application/pdf", svc.gotFile.Tipo)

	body, ct = multipartMaterial(t, false)
	req = httptest.NewRequest("PUT", "/api/contenidos-docentes/c1", body)
	req.Header.Set("Content-Type", ct)
	w = httptest.NewRecorder()
	ctrl.Update(w, conActor(req, "usr-doc", "role_teacher", "c1"))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "c1", svc.got.id)
	assert.Nil(t, svc.gotFile, "editar sin archivo deja el que estaba")

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/contenidos-docentes", strings.NewReader("no es multipart"))
	ctrl.Upload(w, conActor(req, "usr-doc", "role_teacher", ""))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	svc.err = services.NewValidationError("Adjuntá el archivo del material")
	body, ct = multipartMaterial(t, false)
	req = httptest.NewRequest("POST", "/api/contenidos-docentes", body)
	req.Header.Set("Content-Type", ct)
	w = httptest.NewRecorder()
	ctrl.Upload(w, conActor(req, "usr-doc", "role_teacher", ""))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Adjuntá")
}

func TestTeacherContentController_ListadosYModeracion(t *testing.T) {
	svc := &fakeContentService{contenido: models.ContenidoDocente{ID: "c1", DocenteID: "usr-doc", Titulo: "Guía"}}
	ctrl := controllers.NewTeacherContentController(svc)

	w := httptest.NewRecorder()
	ctrl.ListMine(w, conActor(httptest.NewRequest("GET", "/", nil), "usr-doc", "role_teacher", ""))
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	ctrl.List(w, conActor(httptest.NewRequest("GET", "/?estado=pendiente", nil), "a", "role_admin", ""))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pendiente", svc.got.estado)

	w = httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"estado":"rechazado","motivo":"Falta la fuente"}`))
	ctrl.Moderate(w, conActor(req, "usr-admin", "role_admin", "c1"))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "usr-admin", svc.got.actor)
	assert.Equal(t, "Falta la fuente", svc.got.motivo)

	w = httptest.NewRecorder()
	ctrl.Moderate(w, conActor(httptest.NewRequest("POST", "/", strings.NewReader("{")), "a", "role_admin", "c1"))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	ctrl.ListDownloads(w, conActor(httptest.NewRequest("GET", "/", nil), "a", "role_admin", ""))
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	ctrl.ListPublic(w, httptest.NewRequest("GET", "/", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), "usr-doc", "la biblioteca no muestra al docente")

	w = httptest.NewRecorder()
	ctrl.Delete(w, conActor(httptest.NewRequest("DELETE", "/", nil), "usr-doc", "role_teacher", "c1"))
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "role_teacher", svc.got.role)

	svc.err = services.NewForbiddenError("Solo podés eliminar tu propio contenido")
	w = httptest.NewRecorder()
	ctrl.Delete(w, conActor(httptest.NewRequest("DELETE", "/", nil), "otro", "role_teacher", "c1"))
	assert.Equal(t, http.StatusForbidden, w.Code)
	for _, h := range []http.HandlerFunc{ctrl.ListMine, ctrl.List, ctrl.ListDownloads, ctrl.ListPublic} {
		w = httptest.NewRecorder()
		h(w, conActor(httptest.NewRequest("GET", "/", nil), "a", "role_admin", ""))
		assert.Equal(t, http.StatusForbidden, w.Code)
	}
}

func TestTeacherContentController_Descargas(t *testing.T) {
	svc := &fakeContentService{contenido: models.ContenidoDocente{ID: "c1", ArchivoNombre: "guía.pdf", ArchivoTipo: "application/pdf"}}
	ctrl := controllers.NewTeacherContentController(svc)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"nombre":"Ana","apellido":"Gómez","email":"ana@gmail.com","telefono":"3547123456"}`))
	ctrl.PublicDownload(w, conActor(req, "", "", "c1"))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "contenido", w.Body.String())
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
	assert.Equal(t, "Ana", svc.gotDesc.Nombre)

	w = httptest.NewRecorder()
	ctrl.PublicDownload(w, conActor(httptest.NewRequest("POST", "/", strings.NewReader("no")), "", "", "c1"))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	ctrl.File(w, conActor(httptest.NewRequest("GET", "/", nil), "usr-doc", "role_teacher", "c1"))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "usr-doc", svc.got.actor)

	svc.err = repoErrNotFound()
	w = httptest.NewRecorder()
	ctrl.File(w, conActor(httptest.NewRequest("GET", "/", nil), "x", "role_teacher", "c1"))
	assert.Equal(t, http.StatusNotFound, w.Code)
	w = httptest.NewRecorder()
	ctrl.PublicDownload(w, conActor(httptest.NewRequest("POST", "/", strings.NewReader(`{}`)), "", "", "c1"))
	assert.Equal(t, http.StatusNotFound, w.Code)
}
