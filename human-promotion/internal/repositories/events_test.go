package repositories

import (
	"context"
	"errors"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/firestore"

	"buenvivir/human-promotion/internal/audit"
	"buenvivir/human-promotion/internal/events"
	"buenvivir/human-promotion/internal/events/producers"
	"buenvivir/human-promotion/internal/models"
	"buenvivir/human-promotion/internal/outbox"
)

// Tests contra el emulador de Firestore (FIRESTORE_EMULATOR_HOST): el
// cambio, su evento (outbox) y su registro de auditoría van en la misma
// transacción.

func newTestFirestore(t *testing.T) *firestore.Client {
	t.Helper()
	host := os.Getenv("FIRESTORE_EMULATOR_HOST")
	if host == "" {
		t.Skip("FIRESTORE_EMULATOR_HOST no seteado: se omiten los tests contra el emulador de Firestore")
	}
	conn, err := net.DialTimeout("tcp", host, 2*time.Second)
	if err != nil {
		t.Skipf("emulador de Firestore no accesible en %s: %v", host, err)
	}
	_ = conn.Close()
	client, err := firestore.NewClient(context.Background(), "human-promotion-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	for _, c := range []string{fichasCollection, hitosCollection, lineaTiempoCollection, audit.Collection, outbox.Collection} {
		docs, err := client.Collection(c).Documents(context.Background()).GetAll()
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range docs {
			_, _ = d.Ref.Delete(context.Background())
		}
	}
	return client
}

func producer(client *firestore.Client) *producers.PromocionHumana {
	return producers.NewPromocionHumana(outbox.NewWriter(client, events.ProducerPromocionHumana, nil), "promocion-humana-eventos")
}

func requestCtx() context.Context {
	ctx := events.WithActor(context.Background(), "voluntario-1")
	return audit.WithClientIP(ctx, "10.0.0.7")
}

func outboxByType(t *testing.T, client *firestore.Client) map[string][]outbox.Record {
	t.Helper()
	docs, err := client.Collection(outbox.Collection).Documents(context.Background()).GetAll()
	if err != nil {
		t.Fatal(err)
	}
	out := map[string][]outbox.Record{}
	for _, d := range docs {
		var r outbox.Record
		if err := d.DataTo(&r); err != nil {
			t.Fatal(err)
		}
		out[r.EventType] = append(out[r.EventType], r)
	}
	return out
}

func auditEntries(t *testing.T, client *firestore.Client) []audit.Entry {
	t.Helper()
	docs, err := client.Collection(audit.Collection).Documents(context.Background()).GetAll()
	if err != nil {
		t.Fatal(err)
	}
	out := make([]audit.Entry, 0, len(docs))
	for _, d := range docs {
		var e audit.Entry
		_ = d.DataTo(&e)
		out = append(out, e)
	}
	return out
}

func TestFichaYHito_PublicanEventosYAuditan(t *testing.T) {
	client := newTestFirestore(t)
	p := producer(client)
	fichas := NewFichaRepository(client, p)
	hitos := NewHitoRepository(client, p)
	ctx := requestCtx()

	fichaID, err := fichas.Create(ctx, models.FichaDocument{ParticipanteID: "part-1", SituacionActual: "cifrado-1", Avances: "cifrado-2", UpdatedAt: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	if err := fichas.UpdateContenido(ctx, fichaID, "cifrado-3", "cifrado-4", time.Now()); err != nil {
		t.Fatal(err)
	}
	hitoID, err := hitos.Create(ctx, models.HitoDocument{FichaID: fichaID, Fecha: "2026-10-16", Descripcion: "cifrado-5", ResponsableID: "voluntario-1", VisibleParticipante: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := hitos.Update(ctx, hitoID, models.HitoDocument{FichaID: fichaID, Fecha: "2026-10-17", Descripcion: "cifrado-6", ResponsableID: "voluntario-1"}); err != nil {
		t.Fatal(err)
	}

	evs := outboxByType(t, client)
	for _, et := range []string{events.FichaCreada, events.FichaActualizada, events.HitoRegistrado, events.HitoActualizado} {
		if len(evs[et]) != 1 {
			t.Fatalf("%s: %d eventos", et, len(evs[et]))
		}
		r := evs[et][0]
		if r.OrderingKey != "part-1" {
			t.Errorf("%s: ordering key %q", et, r.OrderingKey)
		}
		for _, secreto := range []string{"cifrado", "situacion", "avances", "descripcion"} {
			if containsFold(r.Data, secreto) {
				t.Errorf("%s no debe llevar %q: %s", et, secreto, r.Data)
			}
		}
	}
	env, _ := events.Decode([]byte(evs[events.HitoActualizado][0].Data))
	hito, _ := events.DecodePayload[events.HitoPayload](env)
	if hito.HitoID != hitoID || hito.Fecha != "2026-10-17" || hito.VisibleParticipante || env.ActorID != "voluntario-1" {
		t.Errorf("hito.actualizado: %+v actor %q", hito, env.ActorID)
	}

	entries := auditEntries(t, client)
	if len(entries) != 4 {
		t.Fatalf("auditoría: %d registros", len(entries))
	}
	for _, e := range entries {
		if e.UsuarioID != "voluntario-1" || e.IP != "10.0.0.7" || e.Timestamp.IsZero() || e.RecursoID == "" {
			t.Errorf("registro incompleto: %+v", e)
		}
	}
}

func TestSinProductor_AuditaIgualYNoPublica(t *testing.T) {
	client := newTestFirestore(t)
	fichas := NewFichaRepository(client)
	if _, err := fichas.Create(requestCtx(), models.FichaDocument{ParticipanteID: "part-2"}); err != nil {
		t.Fatal(err)
	}
	if len(outboxByType(t, client)) != 0 {
		t.Error("sin productor no se publica")
	}
	if len(auditEntries(t, client)) != 1 {
		t.Error("la auditoría no depende de Pub/Sub")
	}
}

func TestUpdateInexistente(t *testing.T) {
	client := newTestFirestore(t)
	if err := NewFichaRepository(client).UpdateContenido(requestCtx(), "no-existe", "a", "b", time.Now()); !errors.Is(err, ErrNotFound) {
		t.Errorf("ficha: %v", err)
	}
	if err := NewHitoRepository(client).Update(requestCtx(), "no-existe", models.HitoDocument{FichaID: "f"}); !errors.Is(err, ErrNotFound) {
		t.Errorf("hito: %v", err)
	}
}

func TestDeactivateByParticipante(t *testing.T) {
	client := newTestFirestore(t)
	fichas := NewFichaRepository(client)
	ctx := requestCtx()
	id, err := fichas.Create(ctx, models.FichaDocument{ParticipanteID: "part-3"})
	if err != nil {
		t.Fatal(err)
	}
	if changed, err := fichas.DeactivateByParticipante(ctx, "part-3"); err != nil || !changed {
		t.Fatalf("baja: %v %v", changed, err)
	}
	if changed, err := fichas.DeactivateByParticipante(ctx, "part-3"); err != nil || changed {
		t.Fatalf("repetir la baja no cambia nada: %v %v", changed, err)
	}
	if changed, err := fichas.DeactivateByParticipante(ctx, "sin-ficha"); err != nil || changed {
		t.Fatalf("sin ficha: %v %v", changed, err)
	}
	doc, err := fichas.GetByID(ctx, id)
	if err != nil {
		t.Fatal("la ficha se conserva (baja lógica)")
	}
	if doc.Estado != models.FichaEstadoBaja || doc.FechaBaja == nil {
		t.Errorf("ficha: %+v", doc)
	}
}

func TestLineaTiempoIdempotente(t *testing.T) {
	client := newTestFirestore(t)
	repo := NewLineaTiempoRepository(client)
	e := EntradaLineaTiempo{ParticipanteID: "p1", Tipo: LineaTiempoInscripcion, Descripcion: "Se inscribió en un curso", Fecha: "2026-10-16"}
	for i := 0; i < 2; i++ {
		if err := repo.Add(context.Background(), "evento-1", e); err != nil {
			t.Fatal(err)
		}
	}
	docs, _ := client.Collection(lineaTiempoCollection).Documents(context.Background()).GetAll()
	if len(docs) != 1 {
		t.Fatalf("el mismo evento dos veces deja una sola entrada: %d", len(docs))
	}
}

func containsFold(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}

// ficha.actualizada lleva quién creó la ficha, para que Notificaciones le avise cuando
// otra persona la edita.
func TestFichaActualizada_LlevaCreadoPor(t *testing.T) {
	client := newTestFirestore(t)
	fichas := NewFichaRepository(client, producer(client))
	ctx := requestCtx()

	fichaID, err := fichas.Create(ctx, models.FichaDocument{ParticipanteID: "part-2", CreadoPor: "voluntario-creador", SituacionActual: "c1", Avances: "c2", UpdatedAt: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	if err := fichas.UpdateContenido(events.WithActor(ctx, "admin-editor"), fichaID, "c3", "c4", time.Now()); err != nil {
		t.Fatal(err)
	}
	evs := outboxByType(t, client)
	if len(evs[events.FichaActualizada]) != 1 {
		t.Fatalf("ficha.actualizada: %d eventos", len(evs[events.FichaActualizada]))
	}
	env, err := events.Decode([]byte(evs[events.FichaActualizada][0].Data))
	if err != nil {
		t.Fatal(err)
	}
	p, err := events.DecodePayload[events.FichaPayload](env)
	if err != nil {
		t.Fatal(err)
	}
	if p.CreadoPor != "voluntario-creador" || p.FichaID != fichaID || p.ParticipanteID != "part-2" {
		t.Fatalf("payload de ficha.actualizada: %+v", p)
	}
}
