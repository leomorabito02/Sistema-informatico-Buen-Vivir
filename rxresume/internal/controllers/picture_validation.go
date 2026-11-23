package controllers

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

// Fotos de perfil: además del tope de 5 MB (HU-11) se exige una extensión de imagen y
// que el contenido real lo sea.
var (
	extensionesFoto = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true}
	tiposFoto       = map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true, "image/gif": true}

	errFotoTipo = errors.New("la foto tiene que ser una imagen JPG, PNG, WEBP o GIF")
)

// validarFoto revisa la extensión y los primeros bytes del archivo y deja
// el archivo listo para leerse desde el principio.
func validarFoto(file multipart.File, header *multipart.FileHeader) error {
	if !extensionesFoto[strings.ToLower(filepath.Ext(header.Filename))] {
		return errFotoTipo
	}
	cabecera := make([]byte, 512)
	n, err := io.ReadFull(file, cabecera)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return errFotoTipo
	}
	if !tiposFoto[http.DetectContentType(cabecera[:n])] {
		return errFotoTipo
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	return nil
}
