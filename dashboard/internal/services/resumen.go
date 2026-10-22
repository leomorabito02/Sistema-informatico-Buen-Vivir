package services

import (
	"context"
	"log"
	"sync"

	"buenvivir/dashboard/internal/clients"
	"buenvivir/dashboard/internal/models"
)

// Resumen consulta las cinco fuentes en paralelo (con la caché de svc) y
// calcula indicadores y gráficas. Si una fuente falla, sus indicadores
// salen "No disponible" y el resto se muestra igual: el error solo va al
// log.
func Resumen(ctx context.Context, svc DashboardService, cred clients.Credentials, f Filtros) models.Resumen {
	var d Datos
	var wg sync.WaitGroup
	cargar := func(nombre string, fn func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := fn(); err != nil {
				log.Printf("dashboard resumen: %s no disponible: %v", nombre, err)
			}
		}()
	}
	cargar("participantes", func() error {
		v, err := svc.Participantes(ctx, cred)
		if err == nil {
			d.Participantes = &v
		}
		return err
	})
	cargar("cursos", func() error {
		v, err := svc.Cursos(ctx, cred)
		if err == nil {
			d.Cursos = &v
		}
		return err
	})
	cargar("habilidades", func() error {
		v, err := svc.Habilidades(ctx, cred)
		if err == nil {
			d.Habilidades = &v
		}
		return err
	})
	cargar("postulaciones", func() error {
		v, err := svc.Postulaciones(ctx, cred)
		if err == nil {
			d.Postulaciones = &v
		}
		return err
	})
	cargar("promoción humana", func() error {
		v, err := svc.Promocion(ctx, cred)
		if err == nil {
			d.Promocion = &v
		}
		return err
	})
	wg.Wait()
	return CalcularResumen(d, f)
}
