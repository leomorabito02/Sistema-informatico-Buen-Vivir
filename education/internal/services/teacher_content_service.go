package services

import (
	"context"
	"fmt"
	"io"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
)

const (
	maxTituloContenido      = 120
	minTituloContenido      = 3
	maxDescripcionContenido = 1000
	maxMotivoRechazo        = 500
	maxArchivoContenido     = 10 << 20
	maxEmailDescarga        = 100
)

var soloDigitos = regexp.MustCompile(`^\+?[0-9]{8,15}$`)

// TeacherContentService maneja el contenido de docentes externos (HU-23)
// y la biblioteca pública (HU-25).
type TeacherContentService interface {
	// Del docente.
	Upload(ctx context.Context, docenteID string, in models.ContenidoInput, archivo *models.ArchivoSubido) (*models.ContenidoDocente, error)
	ListMine(ctx context.Context, docenteID string) ([]models.ContenidoDocente, error)
	// Update solo vale para el dueño y mientras el contenido esté
	// pendiente. archivo nil deja el archivo que estaba.
	Update(ctx context.Context, docenteID, id string, in models.ContenidoInput, archivo *models.ArchivoSubido) (*models.ContenidoDocente, error)
	// Delete: el administrador borra cualquiera; el dueño, solo los
	// pendientes.
	Delete(ctx context.Context, actorID, roleID, id string) error

	// De la administración.
	List(ctx context.Context, estado string) ([]models.ContenidoDocente, error)
	Moderate(ctx context.Context, adminID, id, estado, motivo string) (*models.ContenidoDocente, error)
	ListDownloads(ctx context.Context) ([]models.DescargaBiblioteca, error)

	// Biblioteca pública.
	ListPublic(ctx context.Context) ([]models.ContenidoPublico, error)
	// PublicDownload registra quién descarga (HU-25) y devuelve el
	// archivo. Solo para contenido aprobado.
	PublicDownload(ctx context.Context, id string, in models.DescargaInput) (*models.ContenidoDocente, io.ReadCloser, error)

	// File devuelve el archivo a un usuario con sesión: el administrador
	// y el dueño ven cualquiera; el resto, solo los aprobados.
	File(ctx context.Context, actorID, roleID, id string) (*models.ContenidoDocente, io.ReadCloser, error)
}

type teacherContentService struct {
	repo    repositories.TeacherContentRepository
	storage repositories.FileStorage
	now     func() time.Time
}

// NewTeacherContentService recibe el FileStorage de los adjuntos (el
// bucket privado). Sin bucket, subir y descargar responden 503.
func NewTeacherContentService(repo repositories.TeacherContentRepository, storage repositories.FileStorage) TeacherContentService {
	return &teacherContentService{repo: repo, storage: storage, now: time.Now}
}

func esAdministrador(roleID string) bool {
	return roleID == "role_admin" || roleID == "role_dev"
}

func (s *teacherContentService) requireStorage() error {
	if s.storage == nil {
		return NewUnavailableError("La subida de archivos no está configurada en este momento")
	}
	return nil
}

func validarContenidoInput(in models.ContenidoInput) (models.ContenidoInput, error) {
	in.Titulo = SanitizarComentario(in.Titulo)
	in.Descripcion = SanitizarComentario(in.Descripcion)
	if n := len([]rune(in.Titulo)); n < minTituloContenido || n > maxTituloContenido {
		return in, NewValidationError(fmt.Sprintf("El título tiene que tener entre %d y %d caracteres", minTituloContenido, maxTituloContenido))
	}
	if len([]rune(in.Descripcion)) > maxDescripcionContenido {
		return in, NewValidationError(fmt.Sprintf("La descripción puede tener hasta %d caracteres", maxDescripcionContenido))
	}
	return in, nil
}

func validarArchivo(a *models.ArchivoSubido) error {
	if a == nil || len(a.Data) == 0 {
		return NewValidationError("Adjuntá el archivo del material")
	}
	if !allowedSubmissionTypes[a.Tipo] {
		return NewValidationError("El archivo tiene que ser PDF, Word, Excel, PowerPoint, texto o imagen")
	}
	if len(a.Data) > maxArchivoContenido {
		return NewValidationError("El archivo puede pesar hasta 10 MB")
	}
	return nil
}

func (s *teacherContentService) subirArchivo(ctx context.Context, id string, a *models.ArchivoSubido) (string, error) {
	nombre := unsafeFileNameChars.ReplaceAllString(a.Nombre, "_")
	ruta := fmt.Sprintf("contenidos/%s/%d-%s", id, s.now().UnixNano(), nombre)
	if err := s.storage.Upload(ctx, ruta, a.Tipo, a.Data); err != nil {
		return "", fmt.Errorf("error subiendo el archivo: %w", err)
	}
	return ruta, nil
}

func (s *teacherContentService) Upload(ctx context.Context, docenteID string, in models.ContenidoInput, archivo *models.ArchivoSubido) (*models.ContenidoDocente, error) {
	in, err := validarContenidoInput(in)
	if err != nil {
		return nil, err
	}
	if err := validarArchivo(archivo); err != nil {
		return nil, err
	}
	if err := s.requireStorage(); err != nil {
		return nil, err
	}

	id := s.repo.NewID()
	ruta, err := s.subirArchivo(ctx, id, archivo)
	if err != nil {
		return nil, err
	}
	c := &models.ContenidoDocente{
		ID:               id,
		DocenteID:        docenteID,
		Titulo:           in.Titulo,
		Descripcion:      in.Descripcion,
		ArchivoNombre:    archivo.Nombre,
		ArchivoTipo:      archivo.Tipo,
		ArchivoTamano:    int64(len(archivo.Data)),
		ArchivoRuta:      ruta,
		EstadoModeracion: models.ModeracionPendiente,
		FechaSubida:      s.now().UTC(),
	}
	if err := s.repo.Create(ctx, c); err != nil {
		// No dejar el archivo huérfano en el bucket.
		_ = s.storage.DeletePrefix(ctx, "contenidos/"+id+"/")
		return nil, err
	}
	return c, nil
}

func (s *teacherContentService) ListMine(ctx context.Context, docenteID string) ([]models.ContenidoDocente, error) {
	return s.repo.ListByDocente(ctx, docenteID)
}

func (s *teacherContentService) Update(ctx context.Context, docenteID, id string, in models.ContenidoInput, archivo *models.ArchivoSubido) (*models.ContenidoDocente, error) {
	in, err := validarContenidoInput(in)
	if err != nil {
		return nil, err
	}
	c, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if c.DocenteID != docenteID {
		return nil, NewForbiddenError("Solo podés editar tu propio contenido")
	}
	if c.EstadoModeracion != models.ModeracionPendiente {
		return nil, NewValidationError("Solo se puede editar mientras está pendiente de moderación")
	}

	rutaVieja := ""
	if archivo != nil {
		if err := validarArchivo(archivo); err != nil {
			return nil, err
		}
		if err := s.requireStorage(); err != nil {
			return nil, err
		}
		ruta, err := s.subirArchivo(ctx, id, archivo)
		if err != nil {
			return nil, err
		}
		rutaVieja = c.ArchivoRuta
		c.ArchivoNombre, c.ArchivoTipo = archivo.Nombre, archivo.Tipo
		c.ArchivoTamano, c.ArchivoRuta = int64(len(archivo.Data)), ruta
	}
	c.Titulo, c.Descripcion = in.Titulo, in.Descripcion
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	if rutaVieja != "" {
		_ = s.storage.DeletePrefix(ctx, rutaVieja)
	}
	return c, nil
}

func (s *teacherContentService) Delete(ctx context.Context, actorID, roleID, id string) error {
	c, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if !esAdministrador(roleID) {
		if c.DocenteID != actorID {
			return NewForbiddenError("Solo podés eliminar tu propio contenido")
		}
		if c.EstadoModeracion != models.ModeracionPendiente {
			return NewValidationError("Solo se puede eliminar mientras está pendiente de moderación")
		}
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	if s.storage != nil {
		_ = s.storage.DeletePrefix(ctx, "contenidos/"+id+"/")
	}
	return nil
}

func (s *teacherContentService) List(ctx context.Context, estado string) ([]models.ContenidoDocente, error) {
	switch estado {
	case "", models.ModeracionPendiente, models.ModeracionAprobado, models.ModeracionRechazado:
		return s.repo.ListByEstado(ctx, estado)
	}
	return nil, NewValidationError("estado tiene que ser pendiente, aprobado o rechazado")
}

func (s *teacherContentService) Moderate(ctx context.Context, adminID, id, estado, motivo string) (*models.ContenidoDocente, error) {
	if estado != models.ModeracionAprobado && estado != models.ModeracionRechazado {
		return nil, NewValidationError("estado tiene que ser aprobado o rechazado")
	}
	motivo = SanitizarComentario(motivo)
	if estado == models.ModeracionAprobado {
		motivo = ""
	}
	if len([]rune(motivo)) > maxMotivoRechazo {
		return nil, NewValidationError(fmt.Sprintf("El motivo puede tener hasta %d caracteres", maxMotivoRechazo))
	}
	return s.repo.Moderate(ctx, id, estado, motivo, adminID)
}

func (s *teacherContentService) ListDownloads(ctx context.Context) ([]models.DescargaBiblioteca, error) {
	return s.repo.ListDescargas(ctx)
}

func (s *teacherContentService) ListPublic(ctx context.Context) ([]models.ContenidoPublico, error) {
	list, err := s.repo.ListByEstado(ctx, models.ModeracionAprobado)
	if err != nil {
		return nil, err
	}
	out := make([]models.ContenidoPublico, 0, len(list))
	for _, c := range list {
		out = append(out, c.Publico())
	}
	return out, nil
}

// NormalizarTelefono saca espacios, guiones, puntos y paréntesis y
// valida que queden entre 8 y 15 dígitos, con un + opcional adelante
// (HU-25: "validar el formato del número de teléfono").
func NormalizarTelefono(t string) (string, bool) {
	t = strings.Map(func(r rune) rune {
		switch r {
		case ' ', '-', '.', '(', ')':
			return -1
		}
		return r
	}, strings.TrimSpace(t))
	return t, soloDigitos.MatchString(t)
}

func validarDescarga(in models.DescargaInput) (models.DescargaInput, error) {
	in.Nombre = SanitizarComentario(in.Nombre)
	in.Apellido = SanitizarComentario(in.Apellido)
	if in.Nombre == "" || in.Apellido == "" {
		return in, NewValidationError("Completá tu nombre y apellido para descargar")
	}
	if len([]rune(in.Nombre)) > maxNombreComentario || len([]rune(in.Apellido)) > maxNombreComentario {
		return in, NewValidationError("El nombre y el apellido pueden tener hasta 50 caracteres")
	}
	in.Email = strings.TrimSpace(in.Email)
	if addr, err := mail.ParseAddress(in.Email); err != nil || addr.Address != in.Email || len(in.Email) > maxEmailDescarga {
		return in, NewValidationError("El email no es válido")
	}
	tel, ok := NormalizarTelefono(in.Telefono)
	if !ok {
		return in, NewValidationError("El teléfono tiene que tener entre 8 y 15 números, por ejemplo 3547 123456")
	}
	in.Telefono = tel
	return in, nil
}

func (s *teacherContentService) PublicDownload(ctx context.Context, id string, in models.DescargaInput) (*models.ContenidoDocente, io.ReadCloser, error) {
	in, err := validarDescarga(in)
	if err != nil {
		return nil, nil, err
	}
	c, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if c.EstadoModeracion != models.ModeracionAprobado {
		// Lo no aprobado no existe para el público.
		return nil, nil, repositories.ErrNotFound
	}
	if err := s.requireStorage(); err != nil {
		return nil, nil, err
	}
	if err := s.repo.CreateDescarga(ctx, &models.DescargaBiblioteca{
		ContenidoID: c.ID, Nombre: in.Nombre, Apellido: in.Apellido,
		Email: in.Email, Telefono: in.Telefono, Fecha: s.now().UTC(),
	}); err != nil {
		return nil, nil, err
	}
	rc, err := s.storage.Open(ctx, c.ArchivoRuta)
	if err != nil {
		return nil, nil, fmt.Errorf("error abriendo el archivo: %w", err)
	}
	return c, rc, nil
}

func (s *teacherContentService) File(ctx context.Context, actorID, roleID, id string) (*models.ContenidoDocente, io.ReadCloser, error) {
	c, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if !esAdministrador(roleID) && c.DocenteID != actorID && c.EstadoModeracion != models.ModeracionAprobado {
		return nil, nil, NewForbiddenError("Este contenido todavía no está publicado")
	}
	if err := s.requireStorage(); err != nil {
		return nil, nil, err
	}
	rc, err := s.storage.Open(ctx, c.ArchivoRuta)
	if err != nil {
		return nil, nil, fmt.Errorf("error abriendo el archivo: %w", err)
	}
	return c, rc, nil
}
