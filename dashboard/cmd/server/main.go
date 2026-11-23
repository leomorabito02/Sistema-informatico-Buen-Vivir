package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"

	"buenvivir/dashboard/internal/auth"
	"buenvivir/dashboard/internal/clients"
	"buenvivir/dashboard/internal/config"
	"buenvivir/dashboard/internal/controllers"
	"buenvivir/dashboard/internal/events"
	"buenvivir/dashboard/internal/events/consumers"
	"buenvivir/dashboard/internal/observability"
	"buenvivir/dashboard/internal/reportes"
	"buenvivir/dashboard/internal/routes"
	"buenvivir/dashboard/internal/services"
)

func main() {
	observability.InitLogger("dashboard", os.Getenv("APP_ENV"))
	metrics := observability.NewRegistry("dashboard")
	cfg := config.LoadConfig()
	ctx := context.Background()

	var validator *auth.Validator
	if cfg.JWTSecret != "" {
		v, err := auth.NewValidator(cfg.JWTSecret)
		if err != nil {
			log.Fatalf("configuración inválida: %v", err)
		}
		validator = v
	} else if cfg.Production {
		log.Fatalf("configuración inválida: en producción hace falta JWT_SECRET (el gateway de Google Cloud no pasa el rol)")
	}

	upstream := func(nombre, url string) *clients.Upstream {
		u := clients.NewUpstream(nombre, url, cfg.Timeout)
		if cfg.UpstreamIDToken && url != "" {
			u.IDToken = idTokenPara(ctx, url)
		}
		return u
	}
	service := services.NewCachedDashboardService(services.NewDashboardService(
		upstream("autenticación", cfg.AuthURL),
		upstream("educación", cfg.EducationURL),
		upstream("inserción laboral", cfg.LaborURL),
		upstream("promoción humana", cfg.HumanPromotionURL),
	), cfg.CacheTTL)
	controller := controllers.NewDashboardController(service)

	reportesSvc, cerrar, err := reportes.Nuevo(ctx, cfg.ReportesBucket)
	if err != nil {
		log.Fatalf("historial de reportes: %v", err)
	}
	if cerrar != nil {
		defer cerrar()
	} else {
		slog.Info("REPORTES_BUCKET vacío: el historial de reportes no se guarda")
	}

	var eventsHandler http.Handler
	if cfg.ConsumeEnabled() {
		verifier, err := events.NewPushVerifier(cfg.PushAuth, cfg.PushAudience, cfg.PushServiceAccounts, cfg.Production)
		if err != nil {
			log.Fatalf("configuración de Pub/Sub inválida: %v", err)
		}
		eventsHandler = events.CorrelationMiddleware(consumers.NewDashboard(verifier, events.NewMemoryStore(1000), service, slog.Default()))
	}

	mux := http.NewServeMux()
	if observability.MetricsEnabled() {
		mux.Handle("GET /metrics", metrics.Handler())
	}
	routes.RegisterRoutes(mux, routes.Rutas{
		Dashboard: controller,
		Reportes:  reportes.NewHandlers(reportesSvc),
		Eventos:   eventsHandler,
		Guard:     controllers.NewGuard(validator),
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           observability.HTTPMiddleware(metrics, nil)(routes.WithCORS(mux, routes.ParseOrigins(cfg.CORSAllowedOrigins))),
		ReadHeaderTimeout: 10 * time.Second,
		// Cursos con muchas actividades pueden tardar: margen sobre el
		// timeout de cada consulta a un servicio de origen.
		WriteTimeout: 60 * time.Second,
	}
	slog.Info("Servicio Dashboard escuchando", slog.String("port", cfg.Port))
	if err := observability.ListenAndServeGraceful(ctx, srv); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}

// idTokenPara devuelve una función que da un token de identidad de
// Google para el servicio de esa URL. El TokenSource se crea una vez y
// reusa el token mientras no venza.
func idTokenPara(ctx context.Context, audiencia string) func(context.Context) (string, error) {
	var (
		mu sync.Mutex
		ts oauth2.TokenSource
	)
	return func(context.Context) (string, error) {
		mu.Lock()
		defer mu.Unlock()
		if ts == nil {
			s, err := idtoken.NewTokenSource(ctx, audiencia)
			if err != nil {
				return "", err
			}
			ts = s
		}
		tok, err := ts.Token()
		if err != nil {
			return "", err
		}
		return tok.AccessToken, nil
	}
}
