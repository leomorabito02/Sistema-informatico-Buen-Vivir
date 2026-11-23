package services

import (
	"context"
	"testing"

	"buenvivir/labor-integration/api/proto/authpb"
	"buenvivir/labor-integration/internal/models"
)

// fakeAuth responde como Auth: acepta el id de participante o el de
// usuario y devuelve los dos.
type fakeAuth struct{}

func (fakeAuth) GetParticipant(_ context.Context, id string) (*authpb.GetParticipantResponse, error) {
	if id != "part-1" && id != "usr-1" {
		return nil, nil
	}
	return &authpb.GetParticipantResponse{Id: "part-1", UserId: "usr-1", ApprovalStatus: "approved"}, nil
}
func (fakeAuth) GetCompany(context.Context, string) (*authpb.GetCompanyResponse, error) {
	return nil, nil
}
func (fakeAuth) GetActiveUserIDsByRoles(context.Context, []string) ([]string, error) {
	return []string{"admin-1", "vol-1"}, nil
}
func (fakeAuth) GetUserNames(_ context.Context, ids []string) (map[string]string, error) {
	out := map[string]string{}
	for _, id := range ids {
		if id == "usr-1" {
			out[id] = "Paula Participante"
		}
	}
	return out, nil
}
func (fakeAuth) Close() error { return nil }

// fakeApps guarda las postulaciones en memoria.
type fakeApps struct {
	apps []*models.JobApplication
}

func (f *fakeApps) Create(_ context.Context, a *models.JobApplication) error {
	f.apps = append(f.apps, a)
	return nil
}
func (f *fakeApps) ListByOffer(context.Context, string) ([]*models.JobApplication, error) {
	return f.apps, nil
}
func (f *fakeApps) HasApplicationsForOffer(context.Context, string) (bool, error) {
	return len(f.apps) > 0, nil
}
func (f *fakeApps) HasApplied(_ context.Context, offerID string, ids []string) (bool, error) {
	list, _ := f.ListByParticipant(context.Background(), ids)
	for _, a := range list {
		if a.OfferID == offerID {
			return true, nil
		}
	}
	return false, nil
}
func (f *fakeApps) ListByParticipant(_ context.Context, ids []string) ([]*models.JobApplication, error) {
	var out []*models.JobApplication
	for _, a := range f.apps {
		for _, id := range ids {
			if a.ParticipantID == id {
				out = append(out, a)
				break
			}
		}
	}
	return out, nil
}
func (f *fakeApps) UpdateStatus(context.Context, string, string) error { return nil }
func (f *fakeApps) GetByID(_ context.Context, id string) (*models.JobApplication, error) {
	for _, a := range f.apps {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, nil
}
func (f *fakeApps) Delete(context.Context, string) error { return nil }

func TestApplyToJob_GuardaIdDeParticipanteYDeUsuario(t *testing.T) {
	repo := &fakeApps{}
	svc := NewJobApplicationService(repo, nil, fakeAuth{})

	// El participante se postula con el id de usuario del token.
	if err := svc.ApplyToJob(context.Background(), &models.JobApplication{OfferID: "o1", ParticipantID: "usr-1"}); err != nil {
		t.Fatal(err)
	}
	got := repo.apps[0]
	if got.ParticipantID != "part-1" || got.UserID != "usr-1" {
		t.Errorf("guardó participante=%q usuario=%q, quiero part-1 y usr-1", got.ParticipantID, got.UserID)
	}

	// Una segunda postulación a la misma oferta, con cualquiera de los ids, se rechaza.
	if err := svc.ApplyToJob(context.Background(), &models.JobApplication{OfferID: "o1", ParticipantID: "part-1"}); err == nil {
		t.Error("aceptó una postulación repetida")
	}
}

func TestListByParticipant_IncluyePostulacionesViejas(t *testing.T) {
	repo := &fakeApps{apps: []*models.JobApplication{
		{ID: "vieja", OfferID: "o1", ParticipantID: "usr-1"},                   // antes del 29/10: id de usuario
		{ID: "nueva", OfferID: "o2", ParticipantID: "part-1", UserID: "usr-1"}, // ahora
		{ID: "ajena", OfferID: "o1", ParticipantID: "part-2", UserID: "usr-2"},
	}}
	svc := NewJobApplicationService(repo, nil, fakeAuth{})

	apps, err := svc.ListByParticipant(context.Background(), "usr-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(apps) != 2 {
		t.Fatalf("devolvió %d postulaciones, quiero las 2 de usr-1", len(apps))
	}

	// Borrar: el dueño puede con su id de usuario, sea vieja o nueva; otro no.
	for _, id := range []string{"vieja", "nueva"} {
		if err := svc.DeleteApplication(context.Background(), id, "role_participant", "usr-1"); err != nil {
			t.Errorf("no dejó borrar la postulación %s a su dueño: %v", id, err)
		}
	}
	if err := svc.DeleteApplication(context.Background(), "ajena", "role_participant", "usr-1"); err == nil {
		t.Error("dejó borrar una postulación ajena")
	}
}

// Los postulados salen con el nombre de la persona aunque no tenga CV. Vale también
// para las postulaciones viejas con usr-… en participante_id.
func TestListByOffer_CompletaNombreDelPostulante(t *testing.T) {
	repo := &fakeApps{apps: []*models.JobApplication{
		{ID: "a1", ParticipantID: "part-1", UserID: "usr-1", OfferID: "o1"},
		{ID: "a2", ParticipantID: "usr-1", OfferID: "o1"},
		{ID: "a3", ParticipantID: "part-9", UserID: "usr-9", OfferID: "o1"},
	}}
	svc := NewJobApplicationService(repo, nil, fakeAuth{})

	apps, err := svc.ListByOffer(context.Background(), "o1", "role_admin", "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	if apps[0].Nombre != "Paula Participante" || apps[1].Nombre != "Paula Participante" {
		t.Fatalf("nombres esperados, vino %q y %q", apps[0].Nombre, apps[1].Nombre)
	}
	if apps[2].Nombre != "" {
		t.Fatalf("sin nombre en auth debería quedar vacío, vino %q", apps[2].Nombre)
	}
}
