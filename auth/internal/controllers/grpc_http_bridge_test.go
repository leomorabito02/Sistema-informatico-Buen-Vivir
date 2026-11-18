package controllers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"auth/internal/controllers"
	"auth/internal/dto"
	"auth/internal/repositories"
)

func TestGrpcHTTPBridge_Routes(t *testing.T) {
	partSvc := new(MockParticipantService)
	compSvc := new(MockCompanyService)
	userSvc := new(MockUserService)
	devRepo := new(MockDeviceRepository)

	ctrl := controllers.NewGrpcAuthController(partSvc, compSvc, userSvc, nil, devRepo)
	r := chi.NewRouter()
	ctrl.RegisterInternalRoutes(r)

	t.Run("GET /participants/{id}", func(t *testing.T) {
		partSvc.On("GetParticipantByID", mock.Anything, "p-1").Return(&dto.ParticipantResponse{
			ParticipantID: "p-1", UserID: "u-1", ApprovalStatus: "approved",
		}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/participants/p-1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GET /companies/{id}", func(t *testing.T) {
		compSvc.On("GetCompanyByID", mock.Anything, "c-1").Return(&dto.CompanyResponse{
			CompanyID: "c-1", UserID: "u-1", BusinessName: "ACME", Cuit: "30-1", ApprovalStatus: "approved",
		}, nil).Once()
		userSvc.On("GetUserByID", mock.Anything, "u-1").Return(&dto.UserResponse{
			UserID: "u-1", Email: "a@b.c", Phone: "123", Status: "active",
		}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/companies/c-1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("POST /users/active-by-roles", func(t *testing.T) {
		userSvc.On("GetActiveUserIDsByRoles", mock.Anything, []string{"role_admin"}).Return([]string{"u-1"}, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/users/active-by-roles", bytes.NewBufferString(`{"role_ids":["role_admin"]}`))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		badReq := httptest.NewRequest(http.MethodPost, "/users/active-by-roles", bytes.NewBufferString(`{`))
		wBad := httptest.NewRecorder()
		r.ServeHTTP(wBad, badReq)
		assert.Equal(t, http.StatusBadRequest, wBad.Code)
	})

	t.Run("GET and POST device-tokens", func(t *testing.T) {
		devRepo.On("GetUserTokens", mock.Anything, "u-1").Return([]string{"t1"}, nil).Once()
		req := httptest.NewRequest(http.MethodGet, "/users/u-1/device-tokens", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		devRepo.On("DeleteDeviceTokens", mock.Anything, "u-1", []string{"t1"}).Return(nil).Once()
		reqRev := httptest.NewRequest(http.MethodPost, "/users/u-1/device-tokens/revoke", bytes.NewBufferString(`{"tokens":["t1"]}`))
		wRev := httptest.NewRecorder()
		r.ServeHTTP(wRev, reqRev)
		assert.Equal(t, http.StatusOK, wRev.Code)

		reqBad := httptest.NewRequest(http.MethodPost, "/users/u-1/device-tokens/revoke", bytes.NewBufferString(`{`))
		wBad := httptest.NewRecorder()
		r.ServeHTTP(wBad, reqBad)
		assert.Equal(t, http.StatusBadRequest, wBad.Code)
	})

	t.Run("GET /campaigns/recipients, /email-verified, /contact, POST /action-links", func(t *testing.T) {
		userSvc.On("GetCampaignEligibleUsers", mock.Anything).Return([]repositories.CampaignEligibleUser{
			{UserID: "u-1", Email: "a@b.c", IsEmailVerified: true},
		}, nil).Once()
		req := httptest.NewRequest(http.MethodGet, "/campaigns/recipients", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		userSvc.On("IsEmailVerified", mock.Anything, "u-1").Return(true, nil).Once()
		reqVer := httptest.NewRequest(http.MethodGet, "/users/u-1/email-verified", nil)
		wVer := httptest.NewRecorder()
		r.ServeHTTP(wVer, reqVer)
		assert.Equal(t, http.StatusOK, wVer.Code)

		userSvc.On("GetUserByID", mock.Anything, "u-1").Return(&dto.UserResponse{
			UserID: "u-1", Email: "a@b.c", FirstName: "A", LastName: "B", IsEmailVerified: true,
		}, nil).Once()
		reqContact := httptest.NewRequest(http.MethodGet, "/users/u-1/contact", nil)
		wContact := httptest.NewRecorder()
		r.ServeHTTP(wContact, reqContact)
		assert.Equal(t, http.StatusOK, wContact.Code)

		reqLink := httptest.NewRequest(http.MethodPost, "/action-links", bytes.NewBufferString(`{"request_id":"r1","user_id":"u-1","type":"cuenta_aprobada"}`))
		wLink := httptest.NewRecorder()
		r.ServeHTTP(wLink, reqLink)
		assert.Equal(t, http.StatusOK, wLink.Code)

		reqLinkBad := httptest.NewRequest(http.MethodPost, "/action-links", bytes.NewBufferString(`{`))
		wLinkBad := httptest.NewRecorder()
		r.ServeHTTP(wLinkBad, reqLinkBad)
		assert.Equal(t, http.StatusBadRequest, wLinkBad.Code)
	})
}
