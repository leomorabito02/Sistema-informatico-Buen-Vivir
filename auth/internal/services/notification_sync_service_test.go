package services

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"auth/api/proto/notificationpb"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
)

func TestNotificationSyncService_InitUserPreferencesAsync_EmptyUser(t *testing.T) {
	// Arrange
	svc := NewNotificationSyncService()

	// Act & Assert (should not panic or attempt network call)
	svc.InitUserPreferencesAsync("")
	svc.InitUserPreferencesAsync("   ")
}

func TestNotificationSyncService_InitUserPreferencesAsync_NoEnv(t *testing.T) {
	// Arrange
	_ = os.Unsetenv("NOTIFICATIONS_SERVICE_URL")
	svc := NewNotificationSyncService()

	// Act & Assert (should exit immediately without calling)
	svc.InitUserPreferencesAsync("usr-test-123")
}

func TestNotificationSyncService_InitUserPreferencesAsync_Success(t *testing.T) {
	// Arrange
	var receivedPayload map[string]string
	var receivedUserIDHeader string
	var wg sync.WaitGroup
	wg.Add(1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer wg.Done()
		receivedUserIDHeader = r.Header.Get("X-User-ID")
		_ = json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	_ = os.Setenv("NOTIFICATIONS_SERVICE_URL", server.URL)
	defer func() { _ = os.Unsetenv("NOTIFICATIONS_SERVICE_URL") }()

	svc := NewNotificationSyncService()

	// Act
	svc.InitUserPreferencesAsync("usr-test-123")

	// Wait for async execution
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for async notification preference init")
	}

	// Assert
	if receivedUserIDHeader != "usr-test-123" {
		t.Errorf("expected X-User-ID usr-test-123, got %s", receivedUserIDHeader)
	}
	if receivedPayload["user_id"] != "usr-test-123" {
		t.Errorf("expected payload user_id usr-test-123, got %s", receivedPayload["user_id"])
	}
}

type fakeNotificationGRPCServer struct {
	notificationpb.UnimplementedNotificationIngestionServiceServer
	receivedUserID string
	done           chan struct{}
	once           sync.Once
}

func (f *fakeNotificationGRPCServer) InitDefaultPreferences(_ context.Context, req *notificationpb.InitDefaultPreferencesRequest) (*notificationpb.InitDefaultPreferencesResponse, error) {
	f.receivedUserID = req.GetUserId()
	f.once.Do(func() { close(f.done) })
	return &notificationpb.InitDefaultPreferencesResponse{UserId: req.GetUserId(), Success: true, Count: 1}, nil
}

func TestNotificationSyncService_InitUserPreferencesAsync_GRPC(t *testing.T) {
	// Arrange
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	grpcSrv := grpc.NewServer()
	fake := &fakeNotificationGRPCServer{done: make(chan struct{})}
	notificationpb.RegisterNotificationIngestionServiceServer(grpcSrv, fake)
	go func() { _ = grpcSrv.Serve(lis) }()
	t.Cleanup(grpcSrv.Stop)

	_ = os.Setenv("NOTIFICATIONS_GRPC_TARGET", lis.Addr().String())
	defer func() { _ = os.Unsetenv("NOTIFICATIONS_GRPC_TARGET") }()

	svc := NewNotificationSyncService()

	// Act
	svc.InitUserPreferencesAsync("usr-grpc-123")

	// Assert
	select {
	case <-fake.done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for gRPC InitDefaultPreferences")
	}
	if fake.receivedUserID != "usr-grpc-123" {
		t.Errorf("expected user_id usr-grpc-123, got %s", fake.receivedUserID)
	}
}

// Con JWT_SECRET el pedido a Notificaciones lleva un token de servicio
// role_dev firmado (Notificaciones ya no confía en X-User-ID).
func TestServiceToken(t *testing.T) {
	if tok, err := serviceToken(""); err != nil || tok != "" {
		t.Fatalf("sin secreto no debe firmar: %q %v", tok, err)
	}
	tok, err := serviceToken("s3cr3t")
	if err != nil || tok == "" {
		t.Fatalf("no firmó: %v", err)
	}
	parsed, err := jwt.Parse(tok, func(*jwt.Token) (any, error) { return []byte("s3cr3t"), nil })
	if err != nil || !parsed.Valid {
		t.Fatalf("token inválido: %v", err)
	}
	claims := parsed.Claims.(jwt.MapClaims)
	if claims["role"] != "role_dev" || claims["type"] != "access" {
		t.Fatalf("claims inesperados: %v", claims)
	}
}

func TestNotificationSyncService_InitUserPreferencesAsync_SendsServiceHeaders(t *testing.T) {
	// Arrange
	var receivedAuth string
	var receivedServiceAuth string
	var receivedForwardedAuth string
	var wg sync.WaitGroup
	wg.Add(1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer wg.Done()
		receivedAuth = r.Header.Get("Authorization")
		receivedServiceAuth = r.Header.Get("X-Service-Authorization")
		receivedForwardedAuth = r.Header.Get("X-Forwarded-Authorization")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	t.Setenv("NOTIFICATIONS_SERVICE_URL", server.URL)
	t.Setenv("JWT_SECRET", "test-s3cr3t")
	t.Setenv("UPSTREAM_ID_TOKEN", "false")

	svc := NewNotificationSyncService()

	// Act
	svc.InitUserPreferencesAsync("usr-s2s-123")

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for async notification preference init")
	}

	// Assert
	if receivedServiceAuth == "" || receivedForwardedAuth == "" || receivedAuth == "" {
		t.Fatalf("expected service auth headers, got auth=%q svc=%q fwd=%q", receivedAuth, receivedServiceAuth, receivedForwardedAuth)
	}
	if receivedServiceAuth != receivedForwardedAuth {
		t.Fatalf("expected matching X-Service-Authorization and X-Forwarded-Authorization headers")
	}
}
