package services

import (
	"context"
	"regexp"
	"strings"
	"time"
	"unicode"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
)

const (
	maxTextoComentario  = 1000
	maxNombreComentario = 50
	maxEntidadID        = 100
)

var (
	etiquetasHTML = regexp.MustCompile(`<[a-zA-Z/!][^>]*>`)
	entidadIDOK   = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
)

// CommentService maneja los comentarios de cursos y campañas (HU-24).
type CommentService interface {
	List(ctx context.Context, entidadTipo, entidadID string) ([]models.Comentario, error)
	// Create publica un comentario. usuarioID es opcional: el contenido
	// público se comenta sin cuenta.
	Create(ctx context.Context, in models.ComentarioInput, usuarioID string) (*models.Comentario, error)
	// Delete es la moderación del administrador.
	Delete(ctx context.Context, id string) error
}

type commentService struct {
	repo repositories.CommentRepository
	now  func() time.Time
}

func NewCommentService(repo repositories.CommentRepository) CommentService {
	return &commentService{repo: repo, now: time.Now}
}

// normalizarEntidadTipo acepta "campaña" con ñ, que es como lo escribe
// el DER, y lo guarda sin ñ.
func normalizarEntidadTipo(t string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case models.ComentarioCurso:
		return models.ComentarioCurso, nil
	case models.ComentarioCampana, "campaña":
		return models.ComentarioCampana, nil
	}
	return "", NewValidationError("entidadTipo tiene que ser 'curso' o 'campana'")
}

func validarEntidad(tipo, id string) (string, string, error) {
	tipo, err := normalizarEntidadTipo(tipo)
	if err != nil {
		return "", "", err
	}
	id = strings.TrimSpace(id)
	if id == "" || len(id) > maxEntidadID || !entidadIDOK.MatchString(id) {
		return "", "", NewValidationError("entidadId inválido")
	}
	return tipo, id, nil
}

// SanitizarComentario saca las etiquetas HTML y los caracteres de
// control (menos el salto de línea) y recorta espacios: el texto se
// guarda plano, sin nada que un navegador pueda ejecutar (HU-24, XSS).
// Lo que quede de < o > también se saca.
func SanitizarComentario(s string) string {
	s = etiquetasHTML.ReplaceAllString(s, "")
	s = strings.Map(func(r rune) rune {
		if r == '<' || r == '>' {
			return -1
		}
		if unicode.IsControl(r) && r != '\n' {
			return -1
		}
		return r
	}, s)
	return strings.TrimSpace(s)
}

func (s *commentService) List(ctx context.Context, entidadTipo, entidadID string) ([]models.Comentario, error) {
	tipo, id, err := validarEntidad(entidadTipo, entidadID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListByEntidad(ctx, tipo, id)
}

func (s *commentService) Create(ctx context.Context, in models.ComentarioInput, usuarioID string) (*models.Comentario, error) {
	tipo, id, err := validarEntidad(in.EntidadTipo, in.EntidadID)
	if err != nil {
		return nil, err
	}
	nombre := SanitizarComentario(in.Nombre)
	apellido := SanitizarComentario(in.Apellido)
	texto := SanitizarComentario(in.Texto)

	if nombre == "" || apellido == "" {
		return nil, NewValidationError("Completá tu nombre y apellido para comentar")
	}
	if len([]rune(nombre)) > maxNombreComentario || len([]rune(apellido)) > maxNombreComentario {
		return nil, NewValidationError("El nombre y el apellido pueden tener hasta 50 caracteres")
	}
	if texto == "" {
		return nil, NewValidationError("El comentario está vacío")
	}
	if len([]rune(texto)) > maxTextoComentario {
		return nil, NewValidationError("El comentario puede tener hasta 1000 caracteres")
	}

	c := &models.Comentario{
		EntidadTipo:   tipo,
		EntidadID:     id,
		UsuarioID:     usuarioID,
		NombreAutor:   nombre,
		ApellidoAutor: apellido,
		Texto:         texto,
		Fecha:         s.now().UTC(),
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *commentService) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" || !entidadIDOK.MatchString(id) {
		return NewValidationError("id inválido")
	}
	return s.repo.Delete(ctx, id)
}
