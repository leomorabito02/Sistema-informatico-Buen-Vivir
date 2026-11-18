package controllers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func multipartFile(t *testing.T, name string, content []byte) (multipart.File, *multipart.FileHeader) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", name)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = fw.Write(content)
	_ = w.Close()
	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		t.Fatal(err)
	}
	f, h, err := req.FormFile("file")
	if err != nil {
		t.Fatal(err)
	}
	return f, h
}

// HU-11: foto o documento de más de 5 MB se rechaza; uno chico pasa.
func TestValidateUploadedFile_Max5MB(t *testing.T) {
	pdf := append([]byte("%PDF-1.4\n"), bytes.Repeat([]byte("a"), 100)...)
	f, h := multipartFile(t, "dni.pdf", pdf)
	if _, err := validateUploadedFile(f, h); err != nil {
		t.Fatalf("un PDF chico debería pasar: %v", err)
	}

	grande := append([]byte("%PDF-1.4\n"), bytes.Repeat([]byte("a"), MaxUploadBytes+1)...)
	f, h = multipartFile(t, "dni.pdf", grande)
	_, err := validateUploadedFile(f, h)
	if err == nil || !strings.Contains(err.Error(), "5 MB") {
		t.Fatalf("esperaba rechazo por tamaño, dio %v", err)
	}
}
