// Package audit escribe el registro de auditoría de Promoción Humana (HU-35, RNF-14)
// en la colección "auditoria", dentro de la misma transacción que el cambio.
package audit

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"

	"buenvivir/human-promotion/internal/events"
)

// Collection es la colección de registros de auditoría.
const Collection = "auditoria"

// Acciones auditadas.
const (
	AccionCrear      = "crear"
	AccionActualizar = "actualizar"
	AccionBaja       = "baja"
)

// Recursos auditados.
const (
	RecursoFicha = "ficha"
	RecursoHito  = "hito"
)

// Entry es un registro de auditoría. No guarda el contenido de la ficha
// ni del hito (es confidencial): solo qué se tocó.
type Entry struct {
	UsuarioID     string    `firestore:"usuario_id"`
	Accion        string    `firestore:"accion"`
	Recurso       string    `firestore:"recurso"`
	RecursoID     string    `firestore:"recurso_id"`
	Timestamp     time.Time `firestore:"timestamp"`
	IP            string    `firestore:"ip"`
	CorrelationID string    `firestore:"correlation_id"`
}

type ipKey struct{}

// WithClientIP guarda la IP del cliente en el contexto (la pone el
// middleware a partir de RemoteAddr, ya resuelta por middleware.RealIP).
func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, ipKey{}, ip)
}

// ClientIP devuelve la IP guardada con WithClientIP ("" si no hay).
func ClientIP(ctx context.Context) string {
	v, _ := ctx.Value(ipKey{}).(string)
	return v
}

// RecordTx agrega el registro dentro de la transacción tx. El usuario y
// la IP salen del contexto del request; en un cambio originado por un
// evento, el usuario es el actor del evento y la IP queda vacía.
func RecordTx(ctx context.Context, client *firestore.Client, tx *firestore.Transaction, accion, recurso, recursoID string) error {
	entry := Entry{
		UsuarioID:     events.ActorFrom(ctx),
		Accion:        accion,
		Recurso:       recurso,
		RecursoID:     recursoID,
		Timestamp:     time.Now().UTC(),
		IP:            ClientIP(ctx),
		CorrelationID: events.CorrelationIDFrom(ctx),
	}
	if err := tx.Create(client.Collection(Collection).NewDoc(), entry); err != nil {
		return fmt.Errorf("error registrando auditoría: %w", err)
	}
	return nil
}
