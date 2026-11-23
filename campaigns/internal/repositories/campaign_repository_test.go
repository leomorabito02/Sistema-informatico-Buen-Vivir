package repositories

import (
	"context"
	"errors"
	"net"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/firestore"

	"buenvivir/campaigns/internal/models"
)

// Pruebas contra el emulador de Firestore (FIRESTORE_EMULATOR_HOST). Sin
// emulador se saltean, igual que en education y human-promotion.
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
	client, err := firestore.NewClient(context.Background(), "campaigns-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	docs, err := client.Collection(CampaignsCollection).Documents(context.Background()).GetAll()
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range docs {
		if _, err := d.Ref.Delete(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	return client
}

func campana(titulo, fecha, estado string) models.Campaign {
	return models.Campaign{Titulo: titulo, Texto: "texto", ImagenURL: "https://example.org/a.jpg", Fecha: fecha, Estado: estado, CreadoPor: "usuario-1"}
}

func TestCampaignRepository_CrearLeerEditarYPublicar(t *testing.T) {
	repo := NewCampaignRepository(newTestFirestore(t))
	ctx := context.Background()

	creada, err := repo.Create(ctx, campana("Jornada", "2026-11-05", models.EstadoBorrador))
	if err != nil || creada.ID == "" {
		t.Fatalf("create: %+v, %v", creada, err)
	}
	leida, err := repo.Get(ctx, creada.ID)
	if err != nil || leida != creada {
		t.Fatalf("get: %+v, %v (esperaba %+v)", leida, err, creada)
	}

	editada, err := repo.Update(ctx, creada.ID, models.CampaignInput{Titulo: "Jornada nueva", Texto: "otro", ImagenURL: "https://example.org/b.jpg", Fecha: "2026-12-01"})
	if err != nil {
		t.Fatal(err)
	}
	if editada.Titulo != "Jornada nueva" || editada.Fecha != "2026-12-01" || editada.Estado != models.EstadoBorrador || editada.CreadoPor != "usuario-1" {
		t.Fatalf("update cambió algo que no debía: %+v", editada)
	}

	publicada, err := repo.SetEstado(ctx, creada.ID, models.EstadoPublicada)
	if err != nil || publicada.Estado != models.EstadoPublicada || publicada.Titulo != "Jornada nueva" {
		t.Fatalf("set estado: %+v, %v", publicada, err)
	}
}

func TestCampaignRepository_NoExiste(t *testing.T) {
	repo := NewCampaignRepository(newTestFirestore(t))
	ctx := context.Background()
	if _, err := repo.Get(ctx, "noexiste"); !errors.Is(err, ErrNotFound) {
		t.Errorf("get: esperaba ErrNotFound, obtuve %v", err)
	}
	// Editar o publicar una que no existe no la crea.
	if _, err := repo.Update(ctx, "noexiste", models.CampaignInput{Titulo: "x"}); !errors.Is(err, ErrNotFound) {
		t.Errorf("update: esperaba ErrNotFound, obtuve %v", err)
	}
	if _, err := repo.SetEstado(ctx, "noexiste", models.EstadoPublicada); !errors.Is(err, ErrNotFound) {
		t.Errorf("set estado: esperaba ErrNotFound, obtuve %v", err)
	}
	if lista, _ := repo.List(ctx, ""); len(lista) != 0 {
		t.Errorf("se crearon campañas por accidente: %+v", lista)
	}
}

func TestCampaignRepository_ListFiltraYOrdenaPorFecha(t *testing.T) {
	repo := NewCampaignRepository(newTestFirestore(t))
	ctx := context.Background()
	for _, c := range []models.Campaign{
		campana("Vieja", "2026-01-10", models.EstadoPublicada),
		campana("Nueva", "2026-10-10", models.EstadoPublicada),
		campana("Borrador", "2026-12-31", models.EstadoBorrador),
	} {
		if _, err := repo.Create(ctx, c); err != nil {
			t.Fatal(err)
		}
	}

	todas, err := repo.List(ctx, "")
	if err != nil || len(todas) != 3 || todas[0].Titulo != "Borrador" || todas[2].Titulo != "Vieja" {
		t.Fatalf("todas: %+v, %v", todas, err)
	}
	publicadas, err := repo.List(ctx, models.EstadoPublicada)
	if err != nil || len(publicadas) != 2 || publicadas[0].Titulo != "Nueva" || publicadas[1].Titulo != "Vieja" {
		t.Fatalf("publicadas: %+v, %v", publicadas, err)
	}
}

// La baja es lógica: el documento sigue en Firestore, con fecha y usuario, pero no se
// lista ni se lee.
func TestCampaignRepository_BajaLogica(t *testing.T) {
	client := newTestFirestore(t)
	repo := NewCampaignRepository(client)
	ctx := context.Background()

	pub, err := repo.Create(ctx, campana("Publicada", "2026-11-05", models.EstadoPublicada))
	if err != nil {
		t.Fatal(err)
	}
	otra, err := repo.Create(ctx, campana("Otra", "2026-11-06", models.EstadoPublicada))
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(ctx, pub.ID, "admin-1"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	snap, err := client.Collection(CampaignsCollection).Doc(pub.ID).Get(ctx)
	if err != nil {
		t.Fatalf("el documento debería seguir en Firestore: %v", err)
	}
	if snap.Data()["eliminada"] != true || snap.Data()["eliminada_por"] != "admin-1" || snap.Data()["eliminada_en"] == nil {
		t.Fatalf("marca de baja incompleta: %v", snap.Data())
	}
	if _, err := repo.Get(ctx, pub.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("get de una dada de baja: esperaba ErrNotFound, dio %v", err)
	}
	for _, estado := range []string{"", models.EstadoPublicada} {
		lista, err := repo.List(ctx, estado)
		if err != nil {
			t.Fatal(err)
		}
		if len(lista) != 1 || lista[0].ID != otra.ID {
			t.Fatalf("List(%q) debería traer solo la no eliminada: %+v", estado, lista)
		}
	}
	if err := repo.Delete(ctx, pub.ID, "admin-1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("segunda baja: esperaba ErrNotFound, dio %v", err)
	}
	if err := repo.Delete(ctx, "no-existe", "admin-1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("baja de inexistente: esperaba ErrNotFound, dio %v", err)
	}
}
