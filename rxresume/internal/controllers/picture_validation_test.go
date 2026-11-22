package controllers

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"testing"
)

// pngMinimo: firma PNG + cabecera IHDR (alcanza para DetectContentType).
var pngMinimo = append([]byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR"), make([]byte, 32)...)

func archivoMultipart(t *testing.T, nombre string, contenido []byte) (multipart.File, *multipart.FileHeader) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", nombre)
	_, _ = fw.Write(contenido)
	_ = mw.Close()
	req := httptest.NewRequest("POST", "/", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	if err := req.ParseMultipartForm(1 << 20); err != nil {
		t.Fatal(err)
	}
	f, h, err := req.FormFile("file")
	if err != nil {
		t.Fatal(err)
	}
	return f, h
}

func TestValidarFoto(t *testing.T) {
	casos := []struct {
		nombre    string
		contenido []byte
		ok        bool
	}{
		{"foto.png", pngMinimo, true},
		{"FOTO.PNG", pngMinimo, true},
		{"foto.jpg", []byte("\xff\xd8\xff\xe0\x00\x10JFIF\x00"), true},
		{"foto.png", []byte("<html><script>alert(1)</script></html>"), false},
		{"foto.html", pngMinimo, false},
		{"foto.pdf", []byte("%PDF-1.4"), false},
		{"sin-extension", pngMinimo, false},
	}
	for _, c := range casos {
		f, h := archivoMultipart(t, c.nombre, c.contenido)
		err := validarFoto(f, h)
		if (err == nil) != c.ok {
			t.Errorf("%s: ok=%v, err=%v", c.nombre, c.ok, err)
		}
		if c.ok {
			// Tiene que quedar listo para leerse desde el principio.
			primero := make([]byte, 4)
			_, _ = f.Read(primero)
			if !bytes.Equal(primero, c.contenido[:4]) {
				t.Errorf("%s: no volvió al inicio del archivo", c.nombre)
			}
		}
	}
}
