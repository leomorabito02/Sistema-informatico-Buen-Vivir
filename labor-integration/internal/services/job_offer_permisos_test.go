package services

import (
	"context"
	"errors"
	"testing"

	"buenvivir/labor-integration/internal/models"
)

// Auditoría 2026-11-11 (Problema 4): editar, borrar, crear y listar
// ofertas respetan rol y dueño (HU-39 a HU-42).

func svcConOferta(estado string) (*JobOfferService, *mockOfferRepo) {
	repo := newMockOfferRepo()
	repo.offers["off-1"] = &models.JobOffer{ID: "off-1", CompanyID: "usr-empresa", Title: "Pintor", Status: estado}
	apps := &fakeAppsWithGetByID{apps: map[string]*models.JobApplication{}}
	return NewJobOfferService(repo, apps, fakeAuth{}), repo
}

func TestUpdateOffer_PermisosYEstado(t *testing.T) {
	ctx := context.Background()
	cambio := func() *models.JobOffer { return &models.JobOffer{Title: "Pintor (editado)", Status: "Activa"} }

	for _, c := range []struct{ rol, uid string }{
		{"role_participant", "usr-p"}, {"role_teacher", "usr-t"}, {"", ""},
		{"role_company", "usr-otra-empresa"},
	} {
		svc, repo := svcConOferta("Pendiente")
		if err := svc.UpdateOffer(ctx, "off-1", cambio(), c.rol, c.uid); !errors.Is(err, ErrOfferForbidden) {
			t.Errorf("%q/%q: esperaba ErrOfferForbidden, dio %v", c.rol, c.uid, err)
		}
		if repo.offers["off-1"].Title != "Pintor" {
			t.Errorf("%q: la oferta cambió igual", c.rol)
		}
	}

	// La empresa dueña edita, pero no se autoaprueba (HU-40).
	svc, repo := svcConOferta("Pendiente")
	if err := svc.UpdateOffer(ctx, "off-1", cambio(), "role_company", "usr-empresa"); err != nil {
		t.Fatal(err)
	}
	if o := repo.offers["off-1"]; o.Title != "Pintor (editado)" || o.Status != "Pendiente" {
		t.Errorf("empresa dueña: %+v (no debería pasar a Activa)", o)
	}

	// Renovar una vencida sí se permite a la empresa dueña.
	svc, repo = svcConOferta("Vencida")
	if err := svc.UpdateOffer(ctx, "off-1", cambio(), "role_company", "usr-empresa"); err != nil {
		t.Fatal(err)
	}
	if repo.offers["off-1"].Status != "Activa" {
		t.Errorf("renovar: estado %q", repo.offers["off-1"].Status)
	}

	// El voluntario edita cualquiera pero tampoco aprueba.
	svc, repo = svcConOferta("Pendiente")
	if err := svc.UpdateOffer(ctx, "off-1", cambio(), "role_volunteer", "usr-v"); err != nil {
		t.Fatal(err)
	}
	if repo.offers["off-1"].Status != "Pendiente" {
		t.Errorf("voluntario no debería aprobar: %q", repo.offers["off-1"].Status)
	}

	// El admin cambia el estado; un estado vacío no borra el actual.
	svc, repo = svcConOferta("Pendiente")
	if err := svc.UpdateOffer(ctx, "off-1", cambio(), "role_admin", "usr-a"); err != nil || repo.offers["off-1"].Status != "Activa" {
		t.Errorf("admin: %v %q", err, repo.offers["off-1"].Status)
	}
	if err := svc.UpdateOffer(ctx, "off-1", &models.JobOffer{Title: "x"}, "role_admin", "usr-a"); err != nil || repo.offers["off-1"].Status != "Activa" {
		t.Errorf("estado vacío borró el estado: %v %q", err, repo.offers["off-1"].Status)
	}
}

func TestDeleteOffer_Permisos(t *testing.T) {
	ctx := context.Background()
	for _, c := range []struct{ rol, uid string }{
		{"role_participant", "usr-p"}, {"role_teacher", "usr-t"}, {"", ""}, {"role_company", "usr-otra"},
	} {
		svc, repo := svcConOferta("Activa")
		if err := svc.DeleteOffer(ctx, "off-1", c.rol, c.uid); !errors.Is(err, ErrOfferForbidden) {
			t.Errorf("%q: esperaba ErrOfferForbidden, dio %v", c.rol, err)
		}
		if _, ok := repo.offers["off-1"]; !ok {
			t.Errorf("%q: la oferta se borró igual", c.rol)
		}
	}
	for _, c := range []struct{ rol, uid string }{{"role_admin", "a"}, {"role_volunteer", "v"}, {"role_company", "usr-empresa"}} {
		svc, repo := svcConOferta("Activa")
		if err := svc.DeleteOffer(ctx, "off-1", c.rol, c.uid); err != nil {
			t.Errorf("%q: %v", c.rol, err)
		}
		if _, ok := repo.offers["off-1"]; ok {
			t.Errorf("%q: no se borró", c.rol)
		}
	}
}

func TestCreateOffer_SoloEmpresaYGestion(t *testing.T) {
	for _, rol := range []string{"role_participant", "role_teacher", ""} {
		svc, _ := svcConOferta("Activa")
		err := svc.CreateOffer(context.Background(), &models.JobOffer{Title: "x"}, rol, "u")
		if !errors.Is(err, ErrOfferForbidden) {
			t.Errorf("%q: esperaba ErrOfferForbidden, dio %v", rol, err)
		}
	}
}

func TestListOffers_PendientesSoloParaGestion(t *testing.T) {
	repo := newMockOfferRepo()
	repo.offers["a"] = &models.JobOffer{ID: "a", Status: "Activa"}
	repo.offers["p"] = &models.JobOffer{ID: "p", Status: "Pendiente"}
	repo.offers["r"] = &models.JobOffer{ID: "r", Status: "Rechazada"}
	svc := NewJobOfferService(repo, nil, fakeAuth{})
	cuenta := func(rol string) int {
		l, err := svc.ListOffers(context.Background(), "", rol)
		if err != nil {
			t.Fatal(err)
		}
		return len(l)
	}
	for _, rol := range []string{"role_participant", "role_teacher", ""} {
		if n := cuenta(rol); n != 1 {
			t.Errorf("%q ve %d ofertas, esperaba solo la publicada", rol, n)
		}
	}
	for _, rol := range []string{"role_admin", "role_volunteer"} {
		if n := cuenta(rol); n != 3 {
			t.Errorf("%q ve %d ofertas, esperaba 3", rol, n)
		}
	}
}
