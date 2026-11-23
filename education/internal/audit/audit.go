// Package audit escribe el registro de auditoría de Educación (HU-16 y HU-19bis).
// Cada registro se guarda en la misma transacción que el cambio y es de solo lectura.
package audit

import (
	"context"
	"fmt"
	"net"
	"time"

	"cloud.google.com/go/firestore"

	"buenvivir/education/internal/events"
)

const Collection = "auditoria"

const (
	AccionBaja       = "baja"
	AccionRegistrar  = "registrar"
	AccionActualizar = "actualizar"
)

const (
	RecursoCurso             = "curso"
	RecursoAsistenciaGeneral = "asistencia_general"
)

// Entry es un registro de auditoría. No guarda datos personales ni la
// observación de la asistencia: solo qué se tocó.
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

// WithClientIP guarda la IP del cliente en el contexto. Recibe
// RemoteAddr (ya resuelto por middleware.RealIP) y le saca el puerto.
func WithClientIP(ctx context.Context, remoteAddr string) context.Context {
	ip := remoteAddr
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		ip = host
	}
	return context.WithValue(ctx, ipKey{}, ip)
}

// ClientIP devuelve la IP guardada con WithClientIP ("" si no hay).
func ClientIP(ctx context.Context) string {
	v, _ := ctx.Value(ipKey{}).(string)
	return v
}

// NewEntry arma el registro con el usuario, la IP y el correlation id
// del contexto del request.
func NewEntry(ctx context.Context, accion, recurso, recursoID string) Entry {
	return Entry{
		UsuarioID:     events.ActorFrom(ctx),
		Accion:        accion,
		Recurso:       recurso,
		RecursoID:     recursoID,
		Timestamp:     time.Now().UTC(),
		IP:            ClientIP(ctx),
		CorrelationID: events.CorrelationIDFrom(ctx),
	}
}

// RecordTx agrega el registro dentro de la transacción tx.
func RecordTx(ctx context.Context, client *firestore.Client, tx *firestore.Transaction, accion, recurso, recursoID string) error {
	if err := tx.Create(client.Collection(Collection).NewDoc(), NewEntry(ctx, accion, recurso, recursoID)); err != nil {
		return fmt.Errorf("error registrando auditoría: %w", err)
	}
	return nil
}
