package services

import (
	"context"
	"fmt"
	"sort"
	"time"

	"buenvivir/human-promotion/internal/models"
	"buenvivir/human-promotion/internal/repositories"
)

var meses = [...]string{
	"enero", "febrero", "marzo", "abril", "mayo", "junio",
	"julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre",
}

// ActividadService arma la actividad en el centro de un participante a
// partir de la colección linea_tiempo, que llenan los eventos de
// Educación (docs/eventos/README.md, punto 8).
type ActividadService interface {
	// DeParticipante devuelve la actividad en orden cronológico. La
	// asistencia al centro queda en una entrada por mes.
	DeParticipante(ctx context.Context, participanteID string) ([]models.EntradaActividad, error)
	// DeFicha es lo mismo a partir del id de la ficha (vista profesional).
	DeFicha(ctx context.Context, fichaID string) ([]models.EntradaActividad, error)
}

type actividadService struct {
	linea  repositories.LineaTiempoRepository
	fichas repositories.FichaRepository
}

func NewActividadService(linea repositories.LineaTiempoRepository, fichas repositories.FichaRepository) ActividadService {
	return &actividadService{linea: linea, fichas: fichas}
}

func (s *actividadService) DeFicha(ctx context.Context, fichaID string) ([]models.EntradaActividad, error) {
	if !validID(fichaID) {
		return nil, NewValidationError("id de ficha inválido")
	}
	ficha, err := s.fichas.GetByID(ctx, fichaID)
	if err != nil {
		return nil, err
	}
	return s.DeParticipante(ctx, ficha.ParticipanteID)
}

func (s *actividadService) DeParticipante(ctx context.Context, participanteID string) ([]models.EntradaActividad, error) {
	entradas, err := s.linea.ListByParticipante(ctx, participanteID)
	if err != nil {
		return nil, err
	}
	return agruparActividad(entradas), nil
}

// agruparActividad deja una entrada por inscripción o curso completado y
// junta las asistencias del mismo mes en una sola, fechada el primer día
// que asistió. Las entradas con fecha mal formada se descartan.
func agruparActividad(entradas []repositories.EntradaLineaTiempo) []models.EntradaActividad {
	out := []models.EntradaActividad{}
	asistenciaPorMes := map[string]int{} // "AAAA-MM" -> índice en out
	for _, e := range entradas {
		fecha, err := time.Parse(models.FechaLayout, e.Fecha)
		if err != nil {
			continue
		}
		if e.Tipo != repositories.LineaTiempoAsistenciaGeneral {
			out = append(out, models.EntradaActividad{
				ID: e.ID, Tipo: e.Tipo, Fecha: e.Fecha, Descripcion: e.Descripcion, CursoID: e.CursoID,
			})
			continue
		}
		mes := e.Fecha[:7]
		if i, ok := asistenciaPorMes[mes]; ok {
			if e.Fecha < out[i].Fecha {
				out[i].Fecha = e.Fecha
			}
			continue
		}
		asistenciaPorMes[mes] = len(out)
		out = append(out, models.EntradaActividad{
			ID:          "asistencia-" + mes,
			Tipo:        e.Tipo,
			Fecha:       e.Fecha,
			Descripcion: fmt.Sprintf("Asistió al centro en %s de %d", meses[fecha.Month()-1], fecha.Year()),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Fecha != out[j].Fecha {
			return out[i].Fecha < out[j].Fecha
		}
		return out[i].ID < out[j].ID
	})
	return out
}
