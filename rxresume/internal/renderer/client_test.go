package renderer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"buenvivir/rxresume/internal/models"
)

func TestRendererClient_RenderPDF_Success(t *testing.T) {
	// Arrange
	expectedPDF := []byte("%PDF-1.4 mock content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/render" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(expectedPDF)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	resume := &models.Resume{
		ID:     "rsm-1",
		UserID: "usr-1",
		Data: models.ResumeData{
			Basics: models.Basics{Name: "Juan Test"},
		},
	}

	// Act
	pdfBytes, err := client.RenderPDF(context.Background(), resume)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(pdfBytes) != string(expectedPDF) {
		t.Errorf("expected PDF %s, got %s", string(expectedPDF), string(pdfBytes))
	}
}

func TestRendererClient_RenderPDF_ServerError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server failure", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	resume := &models.Resume{ID: "rsm-err"}

	// Act
	_, err := client.RenderPDF(context.Background(), resume)

	// Assert
	if err == nil {
		t.Fatal("expected error from server error status, got nil")
	}
}
