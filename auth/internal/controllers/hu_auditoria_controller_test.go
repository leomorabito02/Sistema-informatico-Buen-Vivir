package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"auth/internal/controllers"
	"auth/internal/dto"
	"auth/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// HU-12 y HU-38: PATCH /users/{id}/approval con rejected rechaza (y el
// servicio pide el email); un valor desconocido es 400.
func TestUserController_Approval_Rejected(t *testing.T) {
	mockSvc := new(MockUserService)
	mockSvc.On("RejectUser", mock.Anything, "usr-1").Return(&dto.UserResponse{UserID: "usr-1", RoleApprovalStatus: "rejected"}, nil)
	r := chi.NewRouter()
	r.Patch("/users/{id}/approval", controllers.NewUserController(mockSvc).ApproveUser)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PATCH", "/users/usr-1/approval", strings.NewReader(`{"approval_status":"rejected"}`)))
	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertCalled(t, "RejectUser", mock.Anything, "usr-1")
	mockSvc.AssertNotCalled(t, "ApproveUser", mock.Anything, mock.Anything)

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("PATCH", "/users/usr-1/approval", strings.NewReader(`{"approval_status":"quizas"}`)))
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// HU-02: la cuenta bloqueada responde 423 con code ACCOUNT_LOCKED.
func TestAuthController_CheckCredentials_Locked(t *testing.T) {
	mockSvc := new(MockAuthService)
	mockSvc.On("CheckCredentials", mock.Anything, mock.Anything).Return(nil, &services.LoginLockedError{Until: time.Now().Add(10 * time.Minute)})
	ctrl := controllers.NewAuthController(mockSvc)

	body, _ := json.Marshal(dto.CheckCredentialsRequest{Email: "ana@gmail.com", Password: "x"})
	w := httptest.NewRecorder()
	ctrl.CheckCredentials(w, httptest.NewRequest("POST", "/api/v1/auth/check-credentials", bytes.NewReader(body)))

	assert.Equal(t, http.StatusLocked, w.Code)
	var resp map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "ACCOUNT_LOCKED", resp["code"])
	assert.Contains(t, resp["error"], "10 minuto")
}
