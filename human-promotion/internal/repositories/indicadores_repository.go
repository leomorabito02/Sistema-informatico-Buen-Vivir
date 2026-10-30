package repositories

import (
	"context"
	"fmt"
	"sort"
	"time"

	"cloud.google.com/go/firestore"

	"buenvivir/human-promotion/internal/audit"
	"buenvivir/human-promotion/internal/models"
)

// IndicadoresRepository junta los datos de los indicadores del Dashboard.
type IndicadoresRepository interface {
	Indicadores(ctx context.Context) (models.Indicadores, error)
}

type indicadoresRepository struct {
	client *firestore.Client
}

func NewIndicadoresRepository(client *firestore.Client) IndicadoresRepository {
	return &indicadoresRepository{client: client}
}

// Indicadores lee fichas e hitos con Select: solo los campos de
// metadatos, sin traer los textos cifrados.
func (r *indicadoresRepository) Indicadores(ctx context.Context) (models.Indicadores, error) {
	fichaDocs, err := r.client.Collection(fichasCollection).
		Select("participante_id", "estado", "fecha_baja", "creada_en", "updated_at").
		Documents(ctx).GetAll()
	if err != nil {
		return models.Indicadores{}, fmt.Errorf("error leyendo fichas: %w", err)
	}

	out := models.Indicadores{Fichas: []models.IndicadorFicha{}, Hitos: []models.IndicadorHito{}}
	participantePorFicha := make(map[string]string, len(fichaDocs))
	sinFecha := map[string]int{} // id de ficha -> posición en out.Fichas
	for _, d := range fichaDocs {
		var doc struct {
			ParticipanteID string     `firestore:"participante_id"`
			Estado         string     `firestore:"estado"`
			FechaBaja      *time.Time `firestore:"fecha_baja"`
			CreadaEn       *time.Time `firestore:"creada_en"`
			UpdatedAt      time.Time  `firestore:"updated_at"`
		}
		if err := d.DataTo(&doc); err != nil {
			return models.Indicadores{}, fmt.Errorf("error leyendo la ficha %s: %w", d.Ref.ID, err)
		}
		f := models.IndicadorFicha{ID: d.Ref.ID, ParticipanteID: doc.ParticipanteID, Estado: models.IndicadorFichaActiva, CreadaEn: doc.UpdatedAt, FechaBaja: doc.FechaBaja}
		if doc.Estado == models.FichaEstadoBaja {
			f.Estado = models.IndicadorFichaBaja
		}
		if doc.CreadaEn != nil {
			f.CreadaEn = *doc.CreadaEn
		} else {
			sinFecha[d.Ref.ID] = len(out.Fichas)
		}
		participantePorFicha[d.Ref.ID] = doc.ParticipanteID
		out.Fichas = append(out.Fichas, f)
	}

	if err := r.completarFechasDeAlta(ctx, out.Fichas, sinFecha); err != nil {
		return models.Indicadores{}, err
	}

	hitoDocs, err := r.client.Collection(hitosCollection).Select("ficha_id", "fecha").Documents(ctx).GetAll()
	if err != nil {
		return models.Indicadores{}, fmt.Errorf("error leyendo hitos: %w", err)
	}
	for _, d := range hitoDocs {
		fichaID, _ := d.Data()["ficha_id"].(string)
		fecha, _ := d.Data()["fecha"].(string)
		participante, ok := participantePorFicha[fichaID]
		if !ok || fecha == "" {
			continue // hito huérfano o sin fecha: no suma a ningún indicador
		}
		out.Hitos = append(out.Hitos, models.IndicadorHito{ParticipanteID: participante, Fecha: fecha})
	}

	sort.Slice(out.Fichas, func(i, j int) bool { return out.Fichas[i].ID < out.Fichas[j].ID })
	sort.Slice(out.Hitos, func(i, j int) bool {
		if out.Hitos[i].Fecha != out.Hitos[j].Fecha {
			return out.Hitos[i].Fecha < out.Hitos[j].Fecha
		}
		return out.Hitos[i].ParticipanteID < out.Hitos[j].ParticipanteID
	})
	return out, nil
}

// completarFechasDeAlta busca en la auditoría el "crear" de las fichas
// que no tienen creada_en (las anteriores a ese campo). Si tampoco está,
// queda la última modificación.
func (r *indicadoresRepository) completarFechasDeAlta(ctx context.Context, fichas []models.IndicadorFicha, sinFecha map[string]int) error {
	if len(sinFecha) == 0 {
		return nil
	}
	docs, err := r.client.Collection(audit.Collection).
		Where("recurso", "==", audit.RecursoFicha).
		Where("accion", "==", audit.AccionCrear).
		Select("recurso_id", "timestamp").
		Documents(ctx).GetAll()
	if err != nil {
		return fmt.Errorf("error leyendo la auditoría: %w", err)
	}
	for _, d := range docs {
		id, _ := d.Data()["recurso_id"].(string)
		ts, _ := d.Data()["timestamp"].(time.Time)
		if i, ok := sinFecha[id]; ok && !ts.IsZero() {
			fichas[i].CreadaEn = ts
		}
	}
	return nil
}
