package repositories

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// AuditCollection guarda la auditoría de Inserción Laboral (HU-42). Nombre
// propio para no mezclarse con la colección "auditoria" de otros servicios
// si comparten la base de Firestore.
const AuditCollection = "auditoria_insercion_laboral"

// AuditEntry es un registro: quién hizo qué sobre qué oferta, cuándo y
// desde qué IP. No guarda el contenido de la oferta. Mismo formato que
// la auditoría de Educación y Promoción Humana.
type AuditEntry struct {
	ID        string    `firestore:"-" json:"id"`
	UsuarioID string    `firestore:"usuario_id" json:"usuario_id"`
	Rol       string    `firestore:"rol" json:"rol,omitempty"`
	Accion    string    `firestore:"accion" json:"accion"`
	Recurso   string    `firestore:"recurso" json:"recurso"`
	RecursoID string    `firestore:"recurso_id" json:"recurso_id"`
	Fecha     time.Time `firestore:"timestamp" json:"fecha"`
	IP        string    `firestore:"ip" json:"ip,omitempty"`
}

// AuditFilter filtra la lectura (Desde inclusive, Hasta exclusivo).
type AuditFilter struct {
	Desde, Hasta time.Time
	Accion       string
}

// AuditRepository solo agrega y lee: no hay edición ni borrado.
type AuditRepository interface {
	Record(ctx context.Context, e AuditEntry) error
	List(ctx context.Context, f AuditFilter) ([]AuditEntry, error)
}

type firestoreAuditRepository struct{ client *firestore.Client }

func NewAuditRepository(client *firestore.Client) AuditRepository {
	return &firestoreAuditRepository{client: client}
}

func (r *firestoreAuditRepository) Record(ctx context.Context, e AuditEntry) error {
	if e.Fecha.IsZero() {
		e.Fecha = time.Now().UTC()
	}
	_, err := r.client.Collection(AuditCollection).NewDoc().Create(ctx, e)
	return err
}

func (r *firestoreAuditRepository) List(ctx context.Context, f AuditFilter) ([]AuditEntry, error) {
	q := r.client.Collection(AuditCollection).Query
	if !f.Desde.IsZero() {
		q = q.Where("timestamp", ">=", f.Desde)
	}
	if !f.Hasta.IsZero() {
		q = q.Where("timestamp", "<", f.Hasta)
	}
	q = q.OrderBy("timestamp", firestore.Desc).Limit(500)
	it := q.Documents(ctx)
	defer it.Stop()
	out := []AuditEntry{}
	for {
		doc, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var e AuditEntry
		if err := doc.DataTo(&e); err != nil {
			return nil, err
		}
		if f.Accion != "" && e.Accion != f.Accion {
			continue
		}
		e.ID = doc.Ref.ID
		out = append(out, e)
	}
	return out, nil
}
