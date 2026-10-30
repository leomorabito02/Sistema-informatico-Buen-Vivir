package repositories

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"

	"buenvivir/human-promotion/internal/audit"
	"buenvivir/human-promotion/internal/models"
)

// AuditoriaRepository lee el registro de auditoría (HU-35). No tiene
// métodos para modificarlo: el log es inmutable y los registros se
// escriben solo dentro de la transacción del cambio (ver paquete audit).
type AuditoriaRepository interface {
	// List devuelve los registros entre desde (incluido) y hasta
	// (excluido), del más nuevo al más viejo, hasta limit.
	List(ctx context.Context, desde, hasta time.Time, limit int) ([]models.RegistroAuditoria, error)
}

type auditoriaRepository struct {
	client *firestore.Client
}

func NewAuditoriaRepository(client *firestore.Client) AuditoriaRepository {
	return &auditoriaRepository{client: client}
}

// List filtra solo por fecha (un único campo: alcanza el índice
// automático de Firestore); el resto de los filtros los aplica el service.
func (r *auditoriaRepository) List(ctx context.Context, desde, hasta time.Time, limit int) ([]models.RegistroAuditoria, error) {
	docs, err := r.client.Collection(audit.Collection).
		Where("timestamp", ">=", desde).
		Where("timestamp", "<", hasta).
		OrderBy("timestamp", firestore.Desc).
		Limit(limit).
		Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("error leyendo la auditoría: %w", err)
	}
	out := make([]models.RegistroAuditoria, 0, len(docs))
	for _, d := range docs {
		var e audit.Entry
		if err := d.DataTo(&e); err != nil {
			return nil, fmt.Errorf("error leyendo el registro %s: %w", d.Ref.ID, err)
		}
		out = append(out, models.RegistroAuditoria{
			ID:        d.Ref.ID,
			UsuarioID: e.UsuarioID,
			Accion:    e.Accion,
			Recurso:   e.Recurso,
			RecursoID: e.RecursoID,
			Fecha:     e.Timestamp,
			IP:        e.IP,
		})
	}
	return out, nil
}
