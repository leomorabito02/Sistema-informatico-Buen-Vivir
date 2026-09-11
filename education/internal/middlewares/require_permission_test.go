package middlewares_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"buenvivir/education/internal/middlewares"
)

// stubPermissionService implementa la interfaz anónima que espera
// RequirePermissionMiddleware. Se copió de auth-service tal cual (no se
// usa todavía en ninguna ruta de education), pero se testea igual para
// no dejarlo sin cobertura y para documentar su contrato.
type stubPermissionService struct {
	allow                           bool
	err                             error
	gotRole, gotAction, gotResource string
}

func (s *stubPermissionService) CheckPermission(_ context.Context, roleID, action, resource string) (bool, error) {
	s.gotRole, s.gotAction, s.gotResource = roleID, action, resource
	return s.allow, s.err
}

func withUserAndRole(userID, roleID string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/cursos", nil)
	ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
	ctx = context.WithValue(ctx, middlewares.RoleIDKey, roleID)
	return req.WithContext(ctx)
}

func TestRequirePermission_NoUserInContext_Unauthorized(t *testing.T) {
	svc := &stubPermissionService{allow: true}
	h := middlewares.RequirePermissionMiddleware(svc, "read", "course")(okHandler())

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/cursos", nil))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequirePermission_NoRoleInContext_Unauthorized(t *testing.T) {
	svc := &stubPermissionService{allow: true}
	h := middlewares.RequirePermissionMiddleware(svc, "read", "course")(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/api/cursos", nil)
	req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDKey, "user-1"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequirePermission_ServiceDenies_Forbidden(t *testing.T) {
	svc := &stubPermissionService{allow: false}
	h := middlewares.RequirePermissionMiddleware(svc, "create", "course")(okHandler())

	w := httptest.NewRecorder()
	h.ServeHTTP(w, withUserAndRole("user-1", "role_participant"))

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, "role_participant", svc.gotRole)
	assert.Equal(t, "create", svc.gotAction)
	assert.Equal(t, "course", svc.gotResource)
}

func TestRequirePermission_ServiceError_Forbidden(t *testing.T) {
	svc := &stubPermissionService{allow: true, err: errors.New("db caída")}
	h := middlewares.RequirePermissionMiddleware(svc, "read", "course")(okHandler())

	w := httptest.NewRecorder()
	h.ServeHTTP(w, withUserAndRole("user-1", "role_admin"))

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequirePermission_ServiceAllows_CallsNext(t *testing.T) {
	svc := &stubPermissionService{allow: true}
	called := false
	h := middlewares.RequirePermissionMiddleware(svc, "read", "course")(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		}),
	)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, withUserAndRole("user-1", "role_admin"))

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
}
