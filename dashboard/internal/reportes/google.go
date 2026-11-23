package reportes

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ── Firestore ──────────────────────────────────────────────────────────

type firestoreRepo struct{ client *firestore.Client }

func NewFirestoreRepository(client *firestore.Client) Repository {
	return &firestoreRepo{client: client}
}

func (r *firestoreRepo) NuevoID() string {
	return r.client.Collection(Collection).NewDoc().ID
}

func (r *firestoreRepo) Crear(ctx context.Context, rep Reporte) error {
	if _, err := r.client.Collection(Collection).Doc(rep.ID).Create(ctx, rep); err != nil {
		return fmt.Errorf("error guardando el reporte: %w", err)
	}
	return nil
}

func (r *firestoreRepo) Obtener(ctx context.Context, id string) (Reporte, error) {
	doc, err := r.client.Collection(Collection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return Reporte{}, ErrNoEncontrado
		}
		return Reporte{}, fmt.Errorf("error leyendo el reporte: %w", err)
	}
	return reporteDe(doc)
}

func (r *firestoreRepo) Listar(ctx context.Context, limite int) ([]Reporte, error) {
	docs, err := r.client.Collection(Collection).OrderBy("fecha_generacion", firestore.Desc).Limit(limite).Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("error listando reportes: %w", err)
	}
	out := make([]Reporte, 0, len(docs))
	for _, d := range docs {
		rep, err := reporteDe(d)
		if err != nil {
			return nil, err
		}
		out = append(out, rep)
	}
	return out, nil
}

func reporteDe(doc *firestore.DocumentSnapshot) (Reporte, error) {
	var rep Reporte
	if err := doc.DataTo(&rep); err != nil {
		return Reporte{}, fmt.Errorf("error leyendo el reporte %s: %w", doc.Ref.ID, err)
	}
	rep.ID = doc.Ref.ID
	return rep, nil
}

// ── Cloud Storage ──────────────────────────────────────────────────────

type gcsStorage struct{ bucket *storage.BucketHandle }

func (s *gcsStorage) Subir(ctx context.Context, objeto, contentType string, data []byte) error {
	w := s.bucket.Object(objeto).If(storage.Conditions{DoesNotExist: true}).NewWriter(ctx)
	w.ContentType = contentType
	if _, err := w.Write(data); err != nil {
		_ = w.Close()
		return fmt.Errorf("error subiendo el PDF: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("error subiendo el PDF: %w", err)
	}
	return nil
}

func (s *gcsStorage) Abrir(ctx context.Context, objeto string) (io.ReadCloser, error) {
	r, err := s.bucket.Object(objeto).NewReader(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil, ErrNoEncontrado
		}
		return nil, fmt.Errorf("error leyendo el PDF: %w", err)
	}
	return r, nil
}

func (s *gcsStorage) Borrar(ctx context.Context, objeto string) error {
	return s.bucket.Object(objeto).Delete(ctx)
}

// ── Arranque ───────────────────────────────────────────────────────────

// Nuevo arma el servicio con Firestore y Cloud Storage. Con bucket vacío devuelve
// (nil, nil, nil): el Dashboard arranca sin historial guardado.
func Nuevo(ctx context.Context, bucket string) (*Service, func(), error) {
	if bucket == "" {
		return nil, nil, nil
	}
	cred, err := credencialesGoogle()
	if err != nil {
		return nil, nil, err
	}
	var opts []option.ClientOption
	if cred != nil {
		opts = append(opts, option.WithCredentialsJSON(cred)) //nolint:staticcheck // credenciales inline, mismo criterio que education
		_ = os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
	}
	var fs *firestore.Client
	if dbID := resolveFirestoreDatabase(); dbID != "" {
		fs, err = firestore.NewClientWithDatabase(ctx, firestore.DetectProjectID, dbID, opts...)
	} else {
		fs, err = firestore.NewClient(ctx, firestore.DetectProjectID, opts...)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("no se pudo inicializar Firestore: %w", err)
	}
	gcs, err := storage.NewClient(ctx, opts...)
	if err != nil {
		_ = fs.Close()
		return nil, nil, fmt.Errorf("no se pudo inicializar Cloud Storage: %w", err)
	}
	cerrar := func() { _ = fs.Close(); _ = gcs.Close() }
	return NewService(NewFirestoreRepository(fs), &gcsStorage{bucket: gcs.Bucket(bucket)}, bucket), cerrar, nil
}

// resolveFirestoreDatabase returns the configured Firestore database ID, or empty string for "(default)".
func resolveFirestoreDatabase() string {
	dbID := strings.TrimSpace(os.Getenv("FIRESTORE_DATABASE"))
	if dbID == "" || dbID == "(default)" || strings.EqualFold(dbID, "null") {
		return ""
	}
	return dbID
}

func credencialesGoogle() ([]byte, error) {
	if b64 := strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_BASE64")); b64 != "" {
		return base64.StdEncoding.DecodeString(b64)
	}
	if v := strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")); strings.HasPrefix(v, "{") {
		return []byte(v), nil
	}
	return nil, nil
}
