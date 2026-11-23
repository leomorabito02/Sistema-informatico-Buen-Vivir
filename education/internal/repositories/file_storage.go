package repositories

import (
	"context"
	"errors"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// ErrFileStorageNotConfigured se devuelve cuando ENTREGAS_BUCKET está
// vacío: el resto de education funciona igual (y se pueden hacer
// entregas solo con texto), pero no se pueden adjuntar archivos.
var ErrFileStorageNotConfigured = errors.New("ENTREGAS_BUCKET no está configurado")

// FileStorage guarda los binarios de los adjuntos de las entregas. Es
// privado: los objetos no tienen URL pública, se leen solo a través de
// education (GET /api/entregas/{id}/archivos/{indice}), que valida
// permisos antes de abrirlos.
type FileStorage interface {
	Upload(ctx context.Context, path, contentType string, data []byte) error
	Open(ctx context.Context, path string) (io.ReadCloser, error)
	// DeletePrefix borra todos los objetos bajo un prefijo (ej. todos
	// los adjuntos de una actividad al eliminarla).
	DeletePrefix(ctx context.Context, prefix string) error
}

type gcsFileStorage struct {
	bucket *storage.BucketHandle
}

// NewGCSFileStorage crea el FileStorage sobre un bucket de Cloud Storage con las mismas
// credenciales de Firestore (nil = ADC). Si bucketName está vacío devuelve (nil, nil).
func NewGCSFileStorage(ctx context.Context, bucketName string, credentialsJSON []byte) (FileStorage, *storage.Client, error) {
	if bucketName == "" {
		return nil, nil, nil
	}
	var opts []option.ClientOption
	if credentialsJSON != nil {
		opts = append(opts, option.WithCredentialsJSON(credentialsJSON)) //nolint:staticcheck // mismo criterio que NewFirestoreClient
	}
	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("no se pudo inicializar Cloud Storage: %w", err)
	}
	return &gcsFileStorage{bucket: client.Bucket(bucketName)}, client, nil
}

func (s *gcsFileStorage) Upload(ctx context.Context, path, contentType string, data []byte) error {
	w := s.bucket.Object(path).If(storage.Conditions{DoesNotExist: true}).NewWriter(ctx)
	w.ContentType = contentType
	if _, err := w.Write(data); err != nil {
		_ = w.Close()
		return fmt.Errorf("error subiendo archivo: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("error subiendo archivo: %w", err)
	}
	return nil
}

func (s *gcsFileStorage) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	r, err := s.bucket.Object(path).NewReader(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("error leyendo archivo: %w", err)
	}
	return r, nil
}

func (s *gcsFileStorage) DeletePrefix(ctx context.Context, prefix string) error {
	it := s.bucket.Objects(ctx, &storage.Query{Prefix: prefix})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error listando archivos: %w", err)
		}
		if err := s.bucket.Object(attrs.Name).Delete(ctx); err != nil && !errors.Is(err, storage.ErrObjectNotExist) {
			return fmt.Errorf("error borrando archivo: %w", err)
		}
	}
}
