package services

import (
	"context"
	"testing"

	"buenvivir/labor-integration/api/proto/authpb"
	"buenvivir/labor-integration/internal/models"
)

func TestIsApprovedStatus(t *testing.T) {
	casos := map[string]bool{
		"approved":  true,
		"accepted":  true,
		" Approved": true,
		"pending":   false,
		"rejected":  false,
		"":          false,
	}
	for in, want := range casos {
		if got := isApprovedStatus(in); got != want {
			t.Errorf("isApprovedStatus(%q) = %v, want %v", in, got, want)
		}
	}
}

// authConAprobacion devuelve una empresa y un participante completos con
// el approval_status pedido.
type authConAprobacion struct {
	fakeAuth
	estado string
}

func (a authConAprobacion) GetCompany(context.Context, string) (*authpb.GetCompanyResponse, error) {
	return &authpb.GetCompanyResponse{
		CompanyId: "comp-1", UserId: "usr-empresa", BusinessName: "Empresa [PRUEBA]",
		Cuit: "30-99999999-7", ApprovalStatus: a.estado,
		UserEmail: "empresa@example.com", UserPhone: "3510000000", UserStatus: "active",
	}, nil
}

func (a authConAprobacion) GetParticipant(context.Context, string) (*authpb.GetParticipantResponse, error) {
	return &authpb.GetParticipantResponse{Id: "part-1", UserId: "usr-1", ApprovalStatus: a.estado}, nil
}

func TestCreateOffer_EmpresaAprobadaConValorViejo(t *testing.T) {
	for _, estado := range []string{"approved", "accepted"} {
		repo := newMockOfferRepo()
		svc := NewJobOfferService(repo, nil, authConAprobacion{estado: estado})
		offer := ofertaEmpresa()
		if err := svc.CreateOffer(context.Background(), offer, "role_company", "usr-empresa"); err != nil {
			t.Fatalf("%q: la empresa aprobada no pudo crear la oferta: %v", estado, err)
		}
		if offer.Status != "Pendiente" {
			t.Errorf("%q: la oferta de empresa debe nacer Pendiente, quedó %q", estado, offer.Status)
		}
	}
	svc := NewJobOfferService(newMockOfferRepo(), nil, authConAprobacion{estado: "pending"})
	if err := svc.CreateOffer(context.Background(), ofertaEmpresa(), "role_company", "usr-empresa"); err == nil {
		t.Error("una empresa pendiente no debería poder crear ofertas")
	}
}

func ofertaEmpresa() *models.JobOffer {
	o := ofertaCompleta("Pintor")
	o.CompanyID = "usr-empresa"
	return o
}
