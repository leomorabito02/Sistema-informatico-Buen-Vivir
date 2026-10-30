// Package integration prueba los flujos de eventos de punta a punta
// contra los emuladores de Firestore y Pub/Sub (si no están, se saltean):
//
//	productor → outbox (Firestore) → relay → Pub/Sub → suscripción con
//	filtro → push → consumidor
//
// El emulador de Pub/Sub no hace push autenticado, así que el test lee
// la suscripción en modo pull y entrega cada mensaje al mismo handler
// HTTP que en producción recibe el push (events.BuildPushBody).
package integration

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"buenvivir/education/internal/controllers"
	"buenvivir/education/internal/events"
	"buenvivir/education/internal/events/consumers"
	"buenvivir/education/internal/events/producers"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/outbox"
	"buenvivir/education/internal/repositories"
)

const project = "education-integration"

func requireEmulator(t *testing.T, env string) {
	t.Helper()
	host := os.Getenv(env)
	if host == "" {
		t.Skipf("%s no seteado: se omite la integración", env)
	}
	conn, err := net.DialTimeout("tcp", host, 2*time.Second)
	if err != nil {
		t.Skipf("emulador %s no accesible: %v", env, err)
	}
	_ = conn.Close()
}

type env struct {
	fs        *firestore.Client
	ps        *pubsub.Client
	publisher *outbox.PubSubPublisher
}

func newEnv(t *testing.T) *env {
	t.Helper()
	requireEmulator(t, "FIRESTORE_EMULATOR_HOST")
	requireEmulator(t, "PUBSUB_EMULATOR_HOST")
	ctx := context.Background()
	fs, err := firestore.NewClient(ctx, project)
	require.NoError(t, err)
	t.Cleanup(func() { _ = fs.Close() })
	for _, c := range []string{outbox.Collection, outbox.ProcessedCollection, "inscripciones"} {
		docs, err := fs.Collection(c).Documents(ctx).GetAll()
		require.NoError(t, err)
		for _, d := range docs {
			_, _ = d.Ref.Delete(ctx)
		}
	}
	ps, err := pubsub.NewClient(ctx, project)
	require.NoError(t, err)
	t.Cleanup(func() { _ = ps.Close() })
	pub, err := outbox.NewPubSubPublisher(ctx, project)
	require.NoError(t, err)
	t.Cleanup(func() { _ = pub.Close() })
	return &env{fs: fs, ps: ps, publisher: pub}
}

// subscription crea (de cero) un topic y una suscripción con el mismo
// filtro por event_type que se usa en producción.
func (e *env) subscription(t *testing.T, topic, sub string, consumer *events.Consumer) string {
	t.Helper()
	ctx := context.Background()
	topicName := "projects/" + project + "/topics/" + topic
	subName := "projects/" + project + "/subscriptions/" + sub
	_ = e.ps.SubscriptionAdminClient.DeleteSubscription(ctx, &pubsubpb.DeleteSubscriptionRequest{Subscription: subName})
	_ = e.ps.TopicAdminClient.DeleteTopic(ctx, &pubsubpb.DeleteTopicRequest{Topic: topicName})
	_, err := e.ps.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{Name: topicName})
	require.NoError(t, err)
	_, err = e.ps.SubscriptionAdminClient.CreateSubscription(ctx, &pubsubpb.Subscription{
		Name: subName, Topic: topicName, AckDeadlineSeconds: 10,
		EnableMessageOrdering: true,
		Filter:                events.SubscriptionFilter(consumer.EventTypes()...),
	})
	require.NoError(t, err)
	return subName
}

// deliver lee hasta `want` mensajes de la suscripción y los entrega al
// consumidor como un push. Confirma (ack) solo si el handler respondió
// 2xx, igual que Pub/Sub.
func (e *env) deliver(t *testing.T, subName string, h http.Handler, want int) []string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var types []string
	err := e.ps.Subscriber(subName).Receive(ctx, func(_ context.Context, m *pubsub.Message) {
		body, err := events.BuildPushBody(m.Data, m.Attributes, m.ID, subName, 1)
		if err != nil {
			m.Nack()
			return
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/eventos", bytes.NewReader(body)))
		if rec.Code >= 300 {
			m.Nack()
			return
		}
		m.Ack()
		types = append(types, m.Attributes[events.AttrEventType])
		if len(types) >= want {
			cancel()
		}
	})
	require.NoError(t, err)
	return types
}

type countingCache struct{ n int }

func (c *countingCache) InvalidateCache() { c.n++ }

// contenido.publicado: webhook de Contentful → outbox → Pub/Sub →
// consumidor de Educación → invalida la caché de contenido.
func TestContenidoPublicado_InvalidaCache(t *testing.T) {
	e := newEnv(t)
	cache := &countingCache{}
	consumer := consumers.NewContenido(events.NoVerifier{}, outbox.NewFirestoreProcessedStore(e.fs), cache, nil)
	sub := e.subscription(t, "contenido-eventos", "educacion-contenido", consumer)

	relay := outbox.NewRelay(e.fs, e.publisher, nil)
	producer := producers.NewEducacion(outbox.NewWriter(e.fs, events.ProducerEducacion, nil), "educacion-eventos", "contenido-eventos")
	webhook := controllers.NewContentfulWebhookController("s3cr3t", producer)

	send := func(topic, body string) int {
		r := httptest.NewRequest(http.MethodPost, "/webhooks/contentful", strings.NewReader(body))
		r.Header.Set(controllers.HeaderWebhookSecret, "s3cr3t")
		r.Header.Set("X-Contentful-Topic", topic)
		rec := httptest.NewRecorder()
		webhook.Receive(rec, r)
		return rec.Code
	}
	body := `{"sys":{"id":"curso-1","revision":3,"contentType":{"sys":{"id":"course"}}}}`
	require.Equal(t, http.StatusAccepted, send("ContentManagement.Entry.publish", body))
	require.Equal(t, http.StatusAccepted, send("ContentManagement.Entry.publish", body), "reintento de Contentful")
	require.Equal(t, http.StatusAccepted, send("ContentManagement.Entry.unpublish", `{"sys":{"id":"curso-1","version":9}}`))

	n, err := relay.Flush(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 2, n, "el reintento del webhook no genera otro evento")

	got := e.deliver(t, sub, consumer, 2)
	assert.ElementsMatch(t, []string{events.ContenidoPublicado, events.ContenidoDespublicado}, got)
	assert.Equal(t, 2, cache.n)
}

// participante.eliminado (lo publicaría Autenticación) → Educación da de
// baja las inscripciones activas del participante y conserva el
// historial. El evento de otro tipo en el mismo topic no le llega
// (filtro de la suscripción).
func TestParticipanteEliminado_DaDeBajaInscripciones(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	enrollments := repositories.NewEnrollmentRepository(e.fs)
	_, err := enrollments.Create(ctx, models.Enrollment{ParticipanteID: "part-9", CursoID: "c1"})
	require.NoError(t, err)

	consumer := consumers.NewUsuarios(events.NoVerifier{}, outbox.NewFirestoreProcessedStore(e.fs), enrollments, nil)
	sub := e.subscription(t, "usuarios-eventos", "educacion-usuarios", consumer)

	// Simula a Autenticación publicando con el envelope estándar.
	publish := func(eventType string, payload any) {
		env, err := events.New(ctx, events.ProducerAutenticacion, eventType, payload)
		require.NoError(t, err)
		data, _ := env.Marshal()
		require.NoError(t, e.publisher.Publish(ctx, "usuarios-eventos", outbox.Message{Data: data, Attributes: env.Attributes(), OrderingKey: "part-9"}))
	}
	publish(events.ParticipanteCreado, events.ParticipantePayload{ParticipanteID: "part-9"}) // Educación no lo consume
	publish(events.ParticipanteEliminado, events.ParticipantePayload{ParticipanteID: "part-9", UserID: "u9"})

	got := e.deliver(t, sub, consumer, 1)
	assert.Equal(t, []string{events.ParticipanteEliminado}, got, "el filtro deja pasar solo lo que Educación consume")

	list, err := enrollments.ListByParticipant(ctx, "part-9")
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, models.EnrollmentEstadoBaja, list[0].Estado)
	assert.Equal(t, models.MotivoBajaParticipanteEliminado, list[0].MotivoBaja)
}
