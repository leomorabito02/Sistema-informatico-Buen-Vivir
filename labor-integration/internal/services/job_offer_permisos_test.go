package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"buenvivir/labor-integration/api/proto/authpb"
	"buenvivir/labor-integration/internal/models"
)

// Editar, borrar, crear y listar ofertas respetan rol y dueño (HU-39 a HU-42).

func svcConOferta(estado string) (*JobOfferService, *mockOfferRepo) {
	repo := newMockOfferRepo()
	repo.offers["off-1"] = &models.JobOffer{ID: "off-1", CompanyID: "usr-empresa", Title: "Pintor", Status: estado}
	apps := &fakeAppsWithGetByID{apps: map[string]*models.JobApplication{}}
	return NewJobOfferService(repo, apps, fakeAuth{}), repo
}

func TestUpdateOffer_PermisosYEstado(t *testing.T) {
	ctx := context.Background()
	cambio := func() *models.JobOffer { o := ofertaCompleta("Pintor (editado)"); o.Status = "Activa"; return o }

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

	// Renovar una vencida o cerrada sí se permite a la empresa dueña.
	for _, estado := range []string{"Vencida", "Cerrada"} {
		svc, repo = svcConOferta(estado)
		now := time.Now()
		repo.offers["off-1"].ClosedAt = &now
		repo.offers["off-1"].ClosedBy = "usr-empresa"
		if err := svc.UpdateOffer(ctx, "off-1", cambio(), "role_company", "usr-empresa"); err != nil {
			t.Fatal(err)
		}
		if repo.offers["off-1"].Status != "Activa" {
			t.Errorf("renovar %s: estado %q", estado, repo.offers["off-1"].Status)
		}
		if repo.offers["off-1"].ClosedAt != nil || repo.offers["off-1"].ClosedBy != "" {
			t.Errorf("renovar %s: no limpió datos de cierre", estado)
		}
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
	if err := svc.UpdateOffer(ctx, "off-1", ofertaCompleta("x"), "role_admin", "usr-a"); err != nil || repo.offers["off-1"].Status != "Activa" {
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
		l, err := svc.ListOffers(context.Background(), "", rol, "")
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

func TestCloseOffer_PermisosYEstado(t *testing.T) {
	ctx := context.Background()

	// Unauthorized roles or other companies are rejected.
	for _, c := range []struct{ rol, uid string }{
		{"role_participant", "usr-p"},
		{"role_teacher", "usr-t"},
		{"", ""},
		{"role_company", "usr-otra-empresa"},
	} {
		svc, repo := svcConOferta("Activa")
		if _, err := svc.CloseOffer(ctx, "off-1", c.rol, c.uid); !errors.Is(err, ErrOfferForbidden) {
			t.Errorf("%q/%q: expected ErrOfferForbidden, got %v", c.rol, c.uid, err)
		}
		if repo.offers["off-1"].Status != "Activa" {
			t.Errorf("%q: offer status changed unexpectedly", c.rol)
		}
	}

	// Non-active offers cannot be closed early.
	for _, st := range []string{"Pendiente", "Vencida", "Cerrada", "Rechazada"} {
		svc, _ := svcConOferta(st)
		if _, err := svc.CloseOffer(ctx, "off-1", "role_company", "usr-empresa"); !errors.Is(err, ErrOfferInvalid) {
			t.Errorf("status %q: expected ErrOfferInvalid, got %v", st, err)
		}
	}

	// Owning company, admin, and volunteer can close an active offer.
	for _, c := range []struct{ rol, uid string }{
		{"role_company", "usr-empresa"},
		{"role_admin", "usr-admin"},
		{"role_volunteer", "usr-vol"},
	} {
		svc, repo := svcConOferta("Activa")
		closed, err := svc.CloseOffer(ctx, "off-1", c.rol, c.uid)
		if err != nil {
			t.Fatalf("%q: unexpected error %v", c.rol, err)
		}
		if closed.Status != "Cerrada" || closed.ClosedBy != c.uid || closed.ClosedAt == nil {
			t.Errorf("%q: unexpected closed offer state %+v", c.rol, closed)
		}
		if repo.offers["off-1"].Status != "Cerrada" {
			t.Errorf("%q: repository status not updated", c.rol)
		}
	}
}

type fakeAuthWithCompany struct{ fakeAuth }

func (fakeAuthWithCompany) GetCompany(_ context.Context, id string) (*authpb.GetCompanyResponse, error) {
	if id != "comp-1" {
		return nil, nil
	}
	return &authpb.GetCompanyResponse{
		CompanyId:    "comp-1",
		UserId:       "comp-1",
		BusinessName: "Panadería Sur",
		UserEmail:    "contacto@panaderiasur.com",
		UserPhone:    "3547112233",
	}, nil
}

func TestListOffers_VisibilityAndCompanyContactByRole(t *testing.T) {
	// Arrange
	repo := newMockOfferRepo()
	repo.offers["act"] = &models.JobOffer{ID: "act", CompanyID: "comp-1", Status: "Activa"}
	repo.offers["ven-app"] = &models.JobOffer{ID: "ven-app", CompanyID: "comp-1", Status: "Vencida"}
	repo.offers["cer-app"] = &models.JobOffer{ID: "cer-app", CompanyID: "comp-1", Status: "Cerrada"}
	repo.offers["cer-no"] = &models.JobOffer{ID: "cer-no", CompanyID: "comp-1", Status: "Cerrada"}
	apps := &fakeApps{apps: []*models.JobApplication{
		{ID: "a1", OfferID: "ven-app", ParticipantID: "part-1", UserID: "usr-1"},
		{ID: "a2", OfferID: "cer-app", ParticipantID: "part-1", UserID: "usr-1"},
	}}
	svc := NewJobOfferService(repo, apps, fakeAuthWithCompany{})

	// Act & Assert: participant sees active + finalized offers they applied to, without company contact
	partOffers, err := svc.ListOffers(context.Background(), "", "role_participant", "usr-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(partOffers) != 3 {
		t.Fatalf("participant expected 3 offers (1 active + 2 applied finalized), got %d", len(partOffers))
	}
	for _, o := range partOffers {
		if o.ID == "cer-no" {
			t.Errorf("participant should not see unapplied closed offer")
		}
		if o.CompanyEmail != "" || o.CompanyPhone != "" {
			t.Errorf("participant should never receive company contact info, got email=%q phone=%q", o.CompanyEmail, o.CompanyPhone)
		}
	}

	// Staff receives company email and phone for follow-up
	for _, staffRole := range []string{"role_admin", "role_volunteer"} {
		staffOffers, err := svc.ListOffers(context.Background(), "", staffRole, "staff-1")
		if err != nil {
			t.Fatal(err)
		}
		if len(staffOffers) != 4 {
			t.Fatalf("%s expected 4 offers, got %d", staffRole, len(staffOffers))
		}
		if staffOffers[0].CompanyEmail != "contacto@panaderiasur.com" || staffOffers[0].CompanyPhone != "3547112233" {
			t.Errorf("%s expected company contact info, got %+v", staffRole, staffOffers[0])
		}
	}
}
