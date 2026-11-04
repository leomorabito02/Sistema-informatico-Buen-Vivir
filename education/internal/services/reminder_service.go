package services

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"buenvivir/education/internal/events"
	"buenvivir/education/internal/events/producers"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
)

// ReminderService triggers activity reminder domain events for scheduled dates (HU-26).
type ReminderService interface {
	TriggerReminders(ctx context.Context, targetDate time.Time) (int, error)
}

type reminderService struct {
	activities  repositories.ActivityRepository
	enrollments repositories.EnrollmentRepository
	producer    *producers.Educacion
}

// NewReminderService builds a new ReminderService instance.
func NewReminderService(
	activities repositories.ActivityRepository,
	enrollments repositories.EnrollmentRepository,
	producer *producers.Educacion,
) ReminderService {
	return &reminderService{
		activities:  activities,
		enrollments: enrollments,
		producer:    producer,
	}
}

// TriggerReminders scans activities falling on targetDate and emits reminders for all active enrollments.
func (s *reminderService) TriggerReminders(ctx context.Context, targetDate time.Time) (int, error) {
	year, month, day := targetDate.UTC().Date()
	start := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, month, day, 23, 59, 59, 999999999, time.UTC)

	acts, err := s.activities.ListByDateRange(ctx, start, end)
	if err != nil {
		return 0, fmt.Errorf("error listando actividades para recordatorios: %w", err)
	}

	count := 0
	for _, act := range acts {
		enrolled, err := s.enrollments.ListByCourse(ctx, act.CursoID)
		if err != nil {
			slog.WarnContext(ctx, "recordatorio: fallo listando inscriptos",
				slog.String("curso_id", act.CursoID),
				slog.String("error", err.Error()),
			)
			continue
		}

		for _, e := range enrolled {
			if e.Estado != models.EnrollmentEstadoActiva {
				continue
			}
			err := s.producer.ActividadRecordatorio(ctx, nil, events.ActividadRecordatorioPayload{
				ActividadID:    act.ID,
				CursoID:        act.CursoID,
				ParticipanteID: e.ParticipanteID,
				Fecha:          act.Fecha.Format("2006-01-02"),
				Titulo:         act.Titulo,
			})
			if err != nil {
				slog.ErrorContext(ctx, "recordatorio: fallo emitiendo evento",
					slog.String("participante_id", e.ParticipanteID),
					slog.String("error", err.Error()),
				)
				continue
			}
			count++
		}
	}

	if count > 0 {
		s.producer.Notify()
	}
	return count, nil
}

// RemindersTriggerHandler handles HTTP POST /internal/reminders/trigger.
func RemindersTriggerHandler(svc ReminderService, verifier events.Verifier) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if verifier != nil {
			if err := verifier.Verify(req); err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		targetDate := time.Now().UTC().AddDate(0, 0, 1) // default: tomorrow
		if rawDate := req.URL.Query().Get("date"); rawDate != "" {
			if parsed, err := time.Parse("2006-01-02", rawDate); err == nil {
				targetDate = parsed
			}
		}

		dispatched, err := svc.TriggerReminders(req.Context(), targetDate)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprintf(w, `{"error":%q}`, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"dispatched":%d}`, dispatched)
	})
}

