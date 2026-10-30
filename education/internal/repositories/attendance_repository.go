package repositories

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"buenvivir/education/internal/models"
)

const attendancesCollection = "asistencias"

// AttendanceRepository encapsula la lectura de la colección Firestore
// "asistencias". Es de solo lectura: las asistencias no se cargan a
// mano, las escribe SubmissionRepository.CreateWithAttendance al
// entregar una actividad.
type AttendanceRepository interface {
	ListByActivity(ctx context.Context, actividadID string) ([]models.Attendance, error)
	ListByParticipant(ctx context.Context, participanteID string) ([]models.Attendance, error)
}

type attendanceRepository struct {
	client *firestore.Client
}

func NewAttendanceRepository(client *firestore.Client) AttendanceRepository {
	return &attendanceRepository{client: client}
}

func (r *attendanceRepository) ListByActivity(ctx context.Context, actividadID string) ([]models.Attendance, error) {
	return r.list(ctx, "actividad_id", actividadID)
}

func (r *attendanceRepository) ListByParticipant(ctx context.Context, participanteID string) ([]models.Attendance, error) {
	return r.list(ctx, "participante_id", participanteID)
}

func (r *attendanceRepository) list(ctx context.Context, field, value string) ([]models.Attendance, error) {
	iter := r.client.Collection(attendancesCollection).Where(field, "==", value).Documents(ctx)
	defer iter.Stop()

	attendances := make([]models.Attendance, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error listando asistencias: %w", err)
		}
		var a models.Attendance
		if err := doc.DataTo(&a); err != nil {
			return nil, fmt.Errorf("error leyendo asistencia: %w", err)
		}
		a.ID = doc.Ref.ID
		attendances = append(attendances, a)
	}
	return attendances, nil
}
