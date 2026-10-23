package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"buenvivir/human-promotion/internal/models"
)

type auditoriaFalsa struct {
	registros    []models.RegistroAuditoria
	desde, hasta time.Time
	err          error
}

func (f *auditoriaFalsa) List(_ context.Context, desde, hasta time.Time, _ int) ([]models.RegistroAuditoria, error) {
	f.desde, f.hasta = desde, hasta
	return f.registros, f.err
}

func TestAuditoria_FiltraYResuelveNombres(t *testing.T) {
	repo := &auditoriaFalsa{registros: []models.RegistroAuditoria{
		{ID: "1", UsuarioID: "u1", Accion: "actualizar", Recurso: "ficha"},
		{ID: "2", UsuarioID: "u2", Accion: "crear", Recurso: "hito"},
	}}
	svc := NewAuditoriaService(repo, authConNombres{usuarios: map[string]string{"u1": "Ana Paz"}})

	got, err := svc.List(context.Background(), FiltroAuditoria{Desde: "2026-10-01", Hasta: "2026-10-22", Recurso: "ficha"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].UsuarioNombre != "Ana Paz" {
		t.Fatalf("resultado inesperado: %+v", got)
	}
	if want := time.Date(2026, 10, 1, 3, 0, 0, 0, time.UTC); !repo.desde.Equal(want) {
		t.Errorf("desde %v, se esperaba %v (medianoche de Argentina)", repo.desde, want)
	}
	if want := time.Date(2026, 10, 23, 3, 0, 0, 0, time.UTC); !repo.hasta.Equal(want) {
		t.Errorf("hasta %v, se esperaba %v (incluye el día hasta)", repo.hasta, want)
	}

	got, err = svc.List(context.Background(), FiltroAuditoria{Accion: "crear"})
	if err != nil || len(got) != 1 || got[0].ID != "2" || got[0].UsuarioNombre != "" {
		t.Fatalf("filtro por acción: %+v %v", got, err)
	}
}

func TestAuditoria_AuthCaidoNoRompe(t *testing.T) {
	repo := &auditoriaFalsa{registros: []models.RegistroAuditoria{{ID: "1", UsuarioID: "u1"}}}
	svc := NewAuditoriaService(repo, authConNombres{err: errors.New("auth caído")})
	got, err := svc.List(context.Background(), FiltroAuditoria{})
	if err != nil || len(got) != 1 {
		t.Fatalf("se esperaba el log sin nombres: %+v %v", got, err)
	}

	repo.err = errors.New("firestore caído")
	if _, err := svc.List(context.Background(), FiltroAuditoria{}); err == nil {
		t.Fatal("se esperaba el error del repositorio")
	}
}

func TestRangoAuditoria(t *testing.T) {
	now := time.Date(2026, 10, 23, 2, 0, 0, 0, time.UTC) // 22/10 23 h en Argentina
	desde, hasta, err := RangoAuditoria("", "", now)
	if err != nil {
		t.Fatal(err)
	}
	if !desde.Equal(time.Date(2026, 9, 23, 3, 0, 0, 0, time.UTC)) || !hasta.Equal(time.Date(2026, 10, 23, 3, 0, 0, 0, time.UTC)) {
		t.Errorf("últimos 30 días: %v a %v", desde, hasta)
	}
	for _, c := range [][2]string{{"mal", ""}, {"", "mal"}, {"2026-10-10", "2026-10-01"}, {"2024-01-01", "2026-01-01"}} {
		if _, _, err := RangoAuditoria(c[0], c[1], now); !IsValidationError(err) {
			t.Errorf("%v: se esperaba error de validación, obtuve %v", c, err)
		}
	}
}
