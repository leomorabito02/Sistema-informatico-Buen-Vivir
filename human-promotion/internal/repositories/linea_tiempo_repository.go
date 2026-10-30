package repositories

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

const lineaTiempoCollection = "linea_tiempo"

// Tipos de entrada de la línea de tiempo que llegan por eventos de
// Educación (HU-33, HU-34, HU-19bis).
const (
	LineaTiempoInscripcion       = "inscripcion"
	LineaTiempoCursoCompletado   = "curso_completado"
	LineaTiempoAsistenciaGeneral = "asistencia_general"
)

// EntradaLineaTiempo es un hecho de la trayectoria del participante que
// registra otro servicio. Sin datos numéricos ni porcentajes: solo qué
// pasó y cuándo (la línea de tiempo la ve el propio participante).
type EntradaLineaTiempo struct {
	// ID es el id del documento (el event_id); no se guarda como campo.
	ID             string    `firestore:"-"`
	ParticipanteID string    `firestore:"participante_id"`
	Tipo           string    `firestore:"tipo"`
	Descripcion    string    `firestore:"descripcion"`
	Fecha          string    `firestore:"fecha"` // AAAA-MM-DD, como los hitos
	CursoID        string    `firestore:"curso_id,omitempty"`
	EventType      string    `firestore:"event_type"`
	CreadoEn       time.Time `firestore:"creado_en"`
}

// LineaTiempoRepository guarda las entradas que llegan por eventos.
type LineaTiempoRepository interface {
	// Add guarda la entrada con id = event_id: si el evento llega dos
	// veces, se pisa la misma entrada (idempotente).
	Add(ctx context.Context, eventID string, e EntradaLineaTiempo) error
	// ListByParticipante devuelve todas las entradas del participante,
	// sin orden (lo ordena quien las muestra).
	ListByParticipante(ctx context.Context, participanteID string) ([]EntradaLineaTiempo, error)
}

type lineaTiempoRepository struct {
	client *firestore.Client
}

func NewLineaTiempoRepository(client *firestore.Client) LineaTiempoRepository {
	return &lineaTiempoRepository{client: client}
}

func (r *lineaTiempoRepository) Add(ctx context.Context, eventID string, e EntradaLineaTiempo) error {
	if _, err := r.client.Collection(lineaTiempoCollection).Doc(eventID).Set(ctx, e); err != nil {
		return fmt.Errorf("error guardando la línea de tiempo: %w", err)
	}
	return nil
}

func (r *lineaTiempoRepository) ListByParticipante(ctx context.Context, participanteID string) ([]EntradaLineaTiempo, error) {
	iter := r.client.Collection(lineaTiempoCollection).Where("participante_id", "==", participanteID).Documents(ctx)
	defer iter.Stop()

	out := []EntradaLineaTiempo{}
	for {
		snap, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error listando la línea de tiempo: %w", err)
		}
		var e EntradaLineaTiempo
		if err := snap.DataTo(&e); err != nil {
			return nil, fmt.Errorf("error leyendo la línea de tiempo: %w", err)
		}
		e.ID = snap.Ref.ID
		out = append(out, e)
	}
	return out, nil
}
