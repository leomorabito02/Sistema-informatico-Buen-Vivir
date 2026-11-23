package routes

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/go-chi/chi/v5"

	"buenvivir/campaigns/internal/controllers"
	"buenvivir/campaigns/internal/middlewares"
	"buenvivir/campaigns/internal/models"
	"buenvivir/campaigns/internal/repositories"
	"buenvivir/campaigns/internal/services"
)

// TestCarga_ListadoPublicoConVolumen es una prueba de carga en proceso:
// servicio real + emulador de Firestore, sin tocar datos reales. Carga
// CARGA_CAMPANAS campañas (500 por defecto) y lanza 50 clientes simultáneos
// (cada uno con su IP) contra el listado público, con y sin búsqueda, y la
// gestión con token. Solo corre con CARGA=1 y el emulador:
//
//	CARGA=1 FIRESTORE_EMULATOR_HOST=127.0.0.1:18390 go test ./internal/routes -run Carga -v
func TestCarga_ListadoPublicoConVolumen(t *testing.T) {
	if os.Getenv("CARGA") != "1" {
		t.Skip("prueba de carga: correr con CARGA=1")
	}
	host := os.Getenv("FIRESTORE_EMULATOR_HOST")
	if host == "" {
		t.Skip("necesita FIRESTORE_EMULATOR_HOST")
	}
	if c, err := net.DialTimeout("tcp", host, 2*time.Second); err != nil {
		t.Skipf("emulador no accesible: %v", err)
	} else {
		_ = c.Close()
	}
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "campaigns-carga")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()
	docs, _ := client.Collection(repositories.CampaignsCollection).Documents(ctx).GetAll()
	for _, d := range docs {
		_, _ = d.Ref.Delete(ctx)
	}

	total := 500
	if v := os.Getenv("CARGA_CAMPANAS"); v != "" {
		_, _ = fmt.Sscanf(v, "%d", &total)
	}
	repo := repositories.NewCampaignRepository(client)
	bw := client.BulkWriter(ctx)
	for i := 0; i < total; i++ {
		estado := models.EstadoPublicada
		if i%3 == 0 {
			estado = models.EstadoBorrador
		}
		c := models.Campaign{Titulo: fmt.Sprintf("Campaña %d de reciclaje y huerta", i), Texto: "texto", ImagenURL: "https://example.org/a.jpg",
			Fecha: fmt.Sprintf("2026-%02d-%02d", 1+i%12, 1+i%28), Estado: estado, CreadoPor: "carga"}
		if _, err := bw.Create(client.Collection(repositories.CampaignsCollection).NewDoc(), c); err != nil {
			t.Fatal(err)
		}
	}
	bw.End()

	r := chi.NewRouter()
	middlewares.SetupMiddlewares(r, "")
	svc := services.NewCampaignService(repo)
	RegisterRoutes(r, services.NewJWTServiceWithSecret(secreto), controllers.NewCampaignController(svc))
	srv := httptest.NewServer(r)
	defer srv.Close()
	vol := token(t, "role_volunteer", "vol-carga")

	type medicion struct {
		d      time.Duration
		status int
	}
	rutas := []struct{ path, tok string }{
		{"/api/v1/campaigns/publicas", ""},
		{"/api/v1/campaigns/publicas?q=huerta%2012", ""},
		{"/api/v1/campaigns?estado=borrador", vol},
	}
	const clientes, porCliente = 50, 20
	var mu sync.Mutex
	var res []medicion
	var wg sync.WaitGroup
	inicio := time.Now()
	for c := 0; c < clientes; c++ {
		wg.Add(1)
		go func(c int) {
			defer wg.Done()
			for i := 0; i < porCliente; i++ {
				ruta := rutas[(c+i)%len(rutas)]
				req, _ := http.NewRequest(http.MethodGet, srv.URL+ruta.path, nil)
				if ruta.tok != "" {
					req.Header.Set("Authorization", "Bearer "+ruta.tok)
				}
				// cada cliente simulado es una persona distinta (su propia
				// IP); si no, el límite de 100/min por IP corta en el pedido 101
				req.Header.Set("X-Real-IP", fmt.Sprintf("10.0.%d.%d", c/250, c%250+1))
				t0 := time.Now()
				resp, err := srv.Client().Do(req)
				st := 0
				if err == nil {
					st = resp.StatusCode
					_ = resp.Body.Close()
				}
				mu.Lock()
				res = append(res, medicion{time.Since(t0), st})
				mu.Unlock()
			}
		}(c)
	}
	wg.Wait()
	dur := time.Since(inicio)

	sort.Slice(res, func(i, j int) bool { return res[i].d < res[j].d })
	pct := func(p float64) time.Duration { return res[int(float64(len(res)-1)*p)].d }
	errores := 0
	porCodigo := map[int]int{}
	for _, m := range res {
		porCodigo[m.status]++
		if m.status != http.StatusOK {
			errores++
		}
	}
	t.Logf("códigos: %v", porCodigo)
	t.Logf("campañas=%d clientes=%d pedidos=%d duración=%s throughput=%.1f req/s p50=%s p95=%s max=%s errores=%d",
		total, clientes, len(res), dur.Round(time.Millisecond), float64(len(res))/dur.Seconds(),
		pct(0.50).Round(time.Millisecond), pct(0.95).Round(time.Millisecond), res[len(res)-1].d.Round(time.Millisecond), errores)
	if errores > 0 {
		t.Errorf("%d pedidos sin 200", errores)
	}
}
