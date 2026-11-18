// Package integration prueba los consumidores de Promoción Humana de
// punta a punta contra los emuladores de Firestore y Pub/Sub (si no
// están, se saltean). El emulador no hace push autenticado: el test lee
// la suscripción en modo pull y entrega cada mensaje al mismo handler
// HTTP que recibe el push en producción.
package integration

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"

	"buenvivir/human-promotion/internal/events"
	"buenvivir/human-promotion/internal/events/consumers"
	"buenvivir/human-promotion/internal/models"
	"buenvivir/human-promotion/internal/outbox"
	"buenvivir/human-promotion/internal/repositories"
)

const project = "human-promotion-integration"

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
	fs  *firestore.Client
	ps  *pubsub.Client
	pub *outbox.PubSubPublisher
}

func newEnv(t *testing.T) *env {
	t.Helper()
	requireEmulator(t, "FIRESTORE_EMULATOR_HOST")
	requireEmulator(t, "PUBSUB_EMULATOR_HOST")
	ctx := context.Background()
	fs, err := firestore.NewClient(ctx, project)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = fs.Close() })
	for _, c := range []string{"linea_tiempo", "fichas_proceso_personal", "auditoria", outbox.ProcessedCollection} {
		docs, _ := fs.Collection(c).Documents(ctx).GetAll()
		for _, d := range docs {
			_, _ = d.Ref.Delete(ctx)
		}
	}
	ps, err := pubsub.NewClient(ctx, project)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ps.Close() })
	pub, err := outbox.NewPubSubPublisher(ctx, project)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pub.Close() })
	return &env{fs: fs, ps: ps, pub: pub}
}

func (e *env) subscription(t *testing.T, topic, sub string, consumer *events.Consumer) string {
	t.Helper()
	ctx := context.Background()
	topicName := "projects/" + project + "/topics/" + topic
	subName := "projects/" + project + "/subscriptions/" + sub
	_ = e.ps.SubscriptionAdminClient.DeleteSubscription(ctx, &pubsubpb.DeleteSubscriptionRequest{Subscription: subName})
	_ = e.ps.TopicAdminClient.DeleteTopic(ctx, &pubsubpb.DeleteTopicRequest{Topic: topicName})
	if _, err := e.ps.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{Name: topicName}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.ps.SubscriptionAdminClient.CreateSubscription(ctx, &pubsubpb.Subscription{
		Name: subName, Topic: topicName, AckDeadlineSeconds: 10, EnableMessageOrdering: true,
		Filter: events.SubscriptionFilter(consumer.EventTypes()...),
	}); err != nil {
		t.Fatal(err)
	}
	return subName
}

func (e *env) publish(t *testing.T, topic, producer, eventType string, payload any, key string) {
	t.Helper()
	ev, err := events.New(context.Background(), producer, eventType, payload)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := ev.Marshal()
	if err := e.pub.Publish(context.Background(), topic, outbox.Message{Data: data, Attributes: ev.Attributes(), OrderingKey: key}); err != nil {
		t.Fatal(err)
	}
}

func (e *env) deliver(t *testing.T, subName string, h http.Handler, want int) []string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var types []string
	err := e.ps.Subscriber(subName).Receive(ctx, func(_ context.Context, m *pubsub.Message) {
		body, _ := events.BuildPushBody(m.Data, m.Attributes, m.ID, subName, 1)
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
	if err != nil {
		t.Fatal(err)
	}
	return types
}

// inscripcion.creada (Educación) → línea de tiempo del participante.
// curso.creado viaja por el mismo topic pero Promoción Humana no lo
// consume: el filtro de la suscripción lo deja afuera.
func TestInscripcionCreada_AgregaLineaDeTiempo(t *testing.T) {
	e := newEnv(t)
	consumer := consumers.NewEducacion(events.NoVerifier{}, outbox.NewFirestoreProcessedStore(e.fs), repositories.NewLineaTiempoRepository(e.fs), nil)
	sub := e.subscription(t, "educacion-eventos", "promocion-humana-educacion", consumer)

	e.publish(t, "educacion-eventos", events.ProducerEducacion, events.CursoCreado, events.CursoPayload{CursoID: "c1"}, "")
	e.publish(t, "educacion-eventos", events.ProducerEducacion, events.InscripcionCreada, events.InscripcionCreadaPayload{
		InscripcionID: "i1", CursoID: "c1", ParticipanteID: "part-1", FechaInscripcion: time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC),
	}, "part-1")

	got := e.deliver(t, sub, consumer, 1)
	if len(got) != 1 || got[0] != events.InscripcionCreada {
		t.Fatalf("llegaron %v", got)
	}
	docs, err := e.fs.Collection("linea_tiempo").Where("participante_id", "==", "part-1").Documents(context.Background()).GetAll()
	if err != nil || len(docs) != 1 {
		t.Fatalf("línea de tiempo: %d docs, %v", len(docs), err)
	}
	var entrada repositories.EntradaLineaTiempo
	_ = docs[0].DataTo(&entrada)
	if entrada.Tipo != repositories.LineaTiempoInscripcion || entrada.Fecha != "2026-10-16" || entrada.CursoID != "c1" {
		t.Errorf("entrada: %+v", entrada)
	}
}

// participante.eliminado (Autenticación) → baja lógica de la ficha.
func TestParticipanteEliminado_BajaLogicaDeFicha(t *testing.T) {
	e := newEnv(t)
	fichas := repositories.NewFichaRepository(e.fs)
	fichaID, err := fichas.Create(context.Background(), models.FichaDocument{ParticipanteID: "part-7"})
	if err != nil {
		t.Fatal(err)
	}
	consumer := consumers.NewUsuarios(events.NoVerifier{}, outbox.NewFirestoreProcessedStore(e.fs), fichas, nil)
	sub := e.subscription(t, "usuarios-eventos", "promocion-humana-usuarios", consumer)

	e.publish(t, "usuarios-eventos", events.ProducerAutenticacion, events.ParticipanteEliminado, events.ParticipantePayload{ParticipanteID: "part-7"}, "part-7")
	if got := e.deliver(t, sub, consumer, 1); len(got) != 1 {
		t.Fatalf("llegaron %v", got)
	}
	doc, err := fichas.GetByID(context.Background(), fichaID)
	if err != nil {
		t.Fatal("la ficha no se borra")
	}
	if doc.Estado != models.FichaEstadoBaja {
		t.Errorf("estado: %q", doc.Estado)
	}
}
