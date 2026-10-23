package repositories

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"

	"buenvivir/education/internal/audit"
	"buenvivir/education/internal/models"
)

// AuditRepository lee el registro de auditoría de Educación. No tiene
// métodos para modificarlo: los registros se escriben solo dentro de la
// transacción del cambio que auditan (ver el paquete audit).
type AuditRepository interface {
	// List devuelve los registros entre desde (incluido) y hasta
	// (excluido), del más nuevo al más viejo, hasta limit.
	List(ctx context.Context, desde, hasta time.Time, limit int) ([]models.RegistroAuditoria, error)
}

type auditRepository struct {
	client *firestore.Client
}

func NewAuditRepository(client *firestore.Client) AuditRepository {
	return &auditRepository{client: client}
}

// List filtra solo por fecha: un único campo, así que alcanza con el
// índice automático de Firestore (los demás filtros los aplica el
// service sobre el resultado).
func (r *auditRepository) List(ctx context.Context, desde, hasta time.Time, limit int) ([]models.RegistroAuditoria, error) {
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
