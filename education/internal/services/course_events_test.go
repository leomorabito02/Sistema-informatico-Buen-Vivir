package services_test

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"buenvivir/education/internal/events"
	"buenvivir/education/internal/events/producers"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/outbox"
	"buenvivir/education/internal/services"
)

// outboxProducer arma un productor real contra el emulador de Firestore
// (o saltea el test si no está).
func outboxProducer(t *testing.T) (*firestore.Client, *producers.Educacion) {
	t.Helper()
	host := os.Getenv("FIRESTORE_EMULATOR_HOST")
	if host == "" {
		t.Skip("FIRESTORE_EMULATOR_HOST no seteado")
	}
	conn, err := net.DialTimeout("tcp", host, 2*time.Second)
	if err != nil {
		t.Skipf("emulador no accesible: %v", err)
	}
	_ = conn.Close()
	client, err := firestore.NewClient(context.Background(), "education-services-test")
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	docs, err := client.Collection(outbox.Collection).Documents(context.Background()).GetAll()
	require.NoError(t, err)
	for _, d := range docs {
		_, _ = d.Ref.Delete(context.Background())
	}
	return client, producers.NewEducacion(outbox.NewWriter(client, events.ProducerEducacion, nil), "educacion-eventos", "contenido-eventos")
}

func outboxTypes(t *testing.T, client *firestore.Client) []string {
	t.Helper()
	docs, err := client.Collection(outbox.Collection).Documents(context.Background()).GetAll()
	require.NoError(t, err)
	var out []string
	for _, d := range docs {
		out = append(out, d.Data()["event_type"].(string))
	}
	return out
}

func TestCourseService_PublicaEventosDeCursos(t *testing.T) {
	client, producer := outboxProducer(t)
	repo := new(MockCourseRepository)
	enrollments := new(MockEnrollmentRepository)
	svc := services.NewCourseServiceConBaja(repo, producer, &fakeBajas{}, enrollments)
	ctx := context.Background()
	input := validCourseInput()
	repo.On("ExistsInManagement", mock.Anything, "c1").Return(true, nil)
	enrollments.On("ListByCourse", mock.Anything, "c1").Return([]models.Enrollment{}, nil)
	course := &models.Course{ID: "c1", ResponsibleTeacher: "doc-1", Mode: "Virtual", TrainingArea: &models.TrainingArea{ID: "a1"}}

	repo.On("Create", mock.Anything, mock.Anything).Return(course, nil)
	repo.On("Update", mock.Anything, "c1", mock.Anything).Return(course, nil)
	repo.On("Publish", mock.Anything, "c1").Return(nil)
	repo.On("Get", mock.Anything, "c1").Return(course, nil)

	_, err := svc.CreateCourse(ctx, input)
	require.NoError(t, err)
	_, err = svc.CreateCourseAsTeacher(ctx, "doc-1", input)
	require.NoError(t, err)
	_, err = svc.UpdateCourse(ctx, "c1", input)
	require.NoError(t, err)
	_, err = svc.UpdateCourseAsTeacher(ctx, "c1", "doc-1", input)
	require.NoError(t, err)
	require.NoError(t, svc.PublishCourse(ctx, "c1"))
	require.NoError(t, svc.DeleteCourse(ctx, "c1"))

	assert.ElementsMatch(t, []string{
		events.CursoCreado, events.CursoCreado, events.ContenidoDocenteSubido,
		events.CursoActualizado, events.CursoActualizado,
		events.ContenidoDocenteModerado, events.CursoEliminado,
	}, outboxTypes(t, client))
}

// Si no se puede registrar el evento, el cambio en Contentful ya está
// hecho: el request no falla (queda en el log).
func TestCourseService_FallaDelEventoNoRompeElRequest(t *testing.T) {
	repo := new(MockCourseRepository)
	broken := producers.NewEducacion(outbox.NewWriter(nil, events.ProducerEducacion, nil), "", "")
	svc := services.NewCourseServiceWithEvents(repo, broken)
	repo.On("Create", mock.Anything, mock.Anything).Return(&models.Course{ID: "c1"}, nil)

	course, err := svc.CreateCourse(context.Background(), validCourseInput())
	require.NoError(t, err)
	assert.Equal(t, "c1", course.ID)
}
