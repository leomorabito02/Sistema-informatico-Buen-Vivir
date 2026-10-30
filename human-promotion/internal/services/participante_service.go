package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"

	"buenvivir/human-promotion/internal/crypto"
	"buenvivir/human-promotion/internal/models"
	"buenvivir/human-promotion/internal/repositories"
)

const maxResultadosBusqueda = 20

// ParticipanteService reúne lo que se consulta a partir del
// participante: la búsqueda por nombre (HU-36) y su propia línea de
// tiempo (HU-34).
type ParticipanteService interface {
	Buscar(ctx context.Context, query string) ([]models.ResultadoBusqueda, error)
	// LineaDeTiempoPropia devuelve solo los hitos marcados como visibles
	// de la ficha del participante logueado, más su actividad en el
	// centro. Si todavía no tiene ficha, los hitos vienen vacíos.
	LineaDeTiempoPropia(ctx context.Context, userID string) (*models.LineaDeTiempoPropia, error)
}

type participanteService struct {
	auth      repositories.AuthDirectory
	fichas    repositories.FichaRepository
	hitos     repositories.HitoRepository
	cipher    crypto.FieldCipher
	actividad ActividadService
}

func NewParticipanteService(auth repositories.AuthDirectory, fichas repositories.FichaRepository, hitos repositories.HitoRepository, cipher crypto.FieldCipher, actividad ActividadService) ParticipanteService {
	return &participanteService{auth: auth, fichas: fichas, hitos: hitos, cipher: cipher, actividad: actividad}
}

// Buscar consulta Auth por nombre y después trae las fichas asociadas.
func (s *participanteService) Buscar(ctx context.Context, query string) ([]models.ResultadoBusqueda, error) {
	query = sanitizeText(query)
	if len([]rune(query)) < 2 {
		return nil, NewValidationError("Ingresá al menos 2 caracteres para buscar")
	}

	encontrados, err := s.auth.SearchParticipantes(ctx, query, maxResultadosBusqueda)
	if errors.Is(err, repositories.ErrInvalidSearch) {
		return nil, NewValidationError("La búsqueda debe tener entre 2 y 100 caracteres")
	}
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(encontrados))
	for _, p := range encontrados {
		ids = append(ids, p.ParticipanteID)
	}
	fichaPorParticipante, err := s.fichas.FindIDsByParticipantes(ctx, ids)
	if err != nil {
		return nil, err
	}

	out := make([]models.ResultadoBusqueda, 0, len(encontrados))
	for _, p := range encontrados {
		r := models.ResultadoBusqueda{
			ParticipanteID: p.ParticipanteID,
			Nombre:         p.Nombre,
			Apellido:       p.Apellido,
		}
		if fichaID, ok := fichaPorParticipante[p.ParticipanteID]; ok {
			r.FichaID = &fichaID
		}
		out = append(out, r)
	}
	return out, nil
}

// LineaDeTiempoPropia filtra en el backend: el participante nunca recibe
// situación actual, avances, responsables ni hitos no visibles (HU-34).
// La actividad en el centro es un dato secundario: si no se puede leer,
// la línea de tiempo sale igual, solo con los hitos.
func (s *participanteService) LineaDeTiempoPropia(ctx context.Context, userID string) (*models.LineaDeTiempoPropia, error) {
	linea := &models.LineaDeTiempoPropia{Hitos: []models.HitoParticipante{}, Actividad: []models.EntradaActividad{}}

	participante, err := s.auth.GetParticipanteByUserID(ctx, userID)
	if errors.Is(err, repositories.ErrNotFound) {
		return linea, nil
	}
	if err != nil {
		return nil, err
	}

	if s.actividad != nil {
		actividad, err := s.actividad.DeParticipante(ctx, participante.ParticipanteID)
		if err != nil {
			slog.Warn("no se pudo leer la actividad del participante", slog.String("error", err.Error()))
		} else {
			linea.Actividad = actividad
		}
	}

	hitos, err := s.hitosVisibles(ctx, participante.ParticipanteID)
	if err != nil {
		return nil, err
	}
	linea.Hitos = hitos
	return linea, nil
}

func (s *participanteService) hitosVisibles(ctx context.Context, participanteID string) ([]models.HitoParticipante, error) {
	vacia := []models.HitoParticipante{}

	fichaID, _, err := s.fichas.GetByParticipante(ctx, participanteID)
	if errors.Is(err, repositories.ErrNotFound) {
		return vacia, nil
	}
	if err != nil {
		return nil, err
	}

	docs, err := s.hitos.ListByFicha(ctx, fichaID)
	if err != nil {
		return nil, err
	}

	out := make([]models.HitoParticipante, 0, len(docs))
	for id, doc := range docs {
		if !doc.VisibleParticipante {
			continue
		}
		descripcion, err := s.cipher.Decrypt(doc.Descripcion)
		if err != nil {
			return nil, fmt.Errorf("error descifrando el hito %s: %w", id, err)
		}
		out = append(out, models.HitoParticipante{ID: id, Fecha: doc.Fecha, Descripcion: descripcion})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Fecha != out[j].Fecha {
			return out[i].Fecha < out[j].Fecha
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}
