package repositories

import (
	"context"
	"fmt"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"buenvivir/education/internal/events/producers"
	"buenvivir/education/internal/models"
)

const (
	teacherContentCollection = "contenidos_docentes"
	libraryDownloadsCol      = "descargas_biblioteca"
)

// TeacherContentRepository encapsula las colecciones Firestore
// "contenidos_docentes" y "descargas_biblioteca" (HU-23, HU-25, DER
// ContenidoDocente y DescargaBiblioteca). Los listados vuelven del más
// nuevo al más viejo.
type TeacherContentRepository interface {
	// NewID reserva el id del documento antes de subir el archivo, para
	// que la ruta del archivo lleve el id del contenido.
	NewID() string
	// Create guarda el contenido con el id reservado y publica
	// contenido_docente.subido en la misma transacción.
	Create(ctx context.Context, c *models.ContenidoDocente) error
	// Get devuelve ErrNotFound si no existe.
	Get(ctx context.Context, id string) (*models.ContenidoDocente, error)
	ListByDocente(ctx context.Context, docenteID string) ([]models.ContenidoDocente, error)
	// ListByEstado con estado vacío devuelve todos.
	ListByEstado(ctx context.Context, estado string) ([]models.ContenidoDocente, error)
	// Update reemplaza título, descripción y archivo. Lo usa la edición
	// del docente mientras el contenido está pendiente.
	Update(ctx context.Context, c *models.ContenidoDocente) error
	// Moderate cambia el estado y publica contenido_docente.moderado en
	// la misma transacción. Devuelve el contenido actualizado.
	Moderate(ctx context.Context, id, estado, motivo, adminID string) (*models.ContenidoDocente, error)
	Delete(ctx context.Context, id string) error

	CreateDescarga(ctx context.Context, d *models.DescargaBiblioteca) error
	ListDescargas(ctx context.Context) ([]models.DescargaBiblioteca, error)
}

type teacherContentRepository struct {
	client *firestore.Client
	events *producers.Educacion
	now    func() time.Time
}

// NewTeacherContentRepository recibe opcionalmente el productor de
// eventos (contenido_docente.subido / .moderado).
func NewTeacherContentRepository(client *firestore.Client, events ...*producers.Educacion) TeacherContentRepository {
	return &teacherContentRepository{client: client, events: firstProducer(events), now: time.Now}
}

func (r *teacherContentRepository) col() *firestore.CollectionRef {
	return r.client.Collection(teacherContentCollection)
}

func (r *teacherContentRepository) NewID() string {
	return r.col().NewDoc().ID
}

func (r *teacherContentRepository) Create(ctx context.Context, c *models.ContenidoDocente) error {
	ref := r.col().Doc(c.ID)
	err := r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		if err := tx.Create(ref, c); err != nil {
			return err
		}
		return r.events.MaterialDocenteSubido(ctx, tx, *c)
	})
	if err != nil {
		return fmt.Errorf("error guardando el contenido: %w", err)
	}
	r.events.Notify()
	return nil
}

func (r *teacherContentRepository) Get(ctx context.Context, id string) (*models.ContenidoDocente, error) {
	doc, err := r.col().Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("error buscando el contenido: %w", err)
	}
	return decodeContenido(doc)
}

func decodeContenido(doc *firestore.DocumentSnapshot) (*models.ContenidoDocente, error) {
	var c models.ContenidoDocente
	if err := doc.DataTo(&c); err != nil {
		return nil, fmt.Errorf("error leyendo el contenido: %w", err)
	}
	c.ID = doc.Ref.ID
	return &c, nil
}

func (r *teacherContentRepository) list(ctx context.Context, q firestore.Query) ([]models.ContenidoDocente, error) {
	iter := q.Documents(ctx)
	defer iter.Stop()
	out := make([]models.ContenidoDocente, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error listando contenidos: %w", err)
		}
		c, err := decodeContenido(doc)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	// Solo filtros de igualdad: el orden se hace acá para no necesitar
	// un índice compuesto en Firestore.
	sort.SliceStable(out, func(i, j int) bool { return out[i].FechaSubida.After(out[j].FechaSubida) })
	return out, nil
}

func (r *teacherContentRepository) ListByDocente(ctx context.Context, docenteID string) ([]models.ContenidoDocente, error) {
	return r.list(ctx, r.col().Where("docente_id", "==", docenteID))
}

func (r *teacherContentRepository) ListByEstado(ctx context.Context, estado string) ([]models.ContenidoDocente, error) {
	if estado == "" {
		return r.list(ctx, r.col().Query)
	}
	return r.list(ctx, r.col().Where("estado_moderacion", "==", estado))
}

func (r *teacherContentRepository) Update(ctx context.Context, c *models.ContenidoDocente) error {
	_, err := r.col().Doc(c.ID).Update(ctx, []firestore.Update{
		{Path: "titulo", Value: c.Titulo},
		{Path: "descripcion", Value: c.Descripcion},
		{Path: "archivo_nombre", Value: c.ArchivoNombre},
		{Path: "archivo_tipo", Value: c.ArchivoTipo},
		{Path: "archivo_tamano", Value: c.ArchivoTamano},
		{Path: "archivo_ruta", Value: c.ArchivoRuta},
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return ErrNotFound
		}
		return fmt.Errorf("error actualizando el contenido: %w", err)
	}
	return nil
}

func (r *teacherContentRepository) Moderate(ctx context.Context, id, estado, motivo, adminID string) (*models.ContenidoDocente, error) {
	ref := r.col().Doc(id)
	var out *models.ContenidoDocente
	err := r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(ref)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return ErrNotFound
			}
			return err
		}
		c, err := decodeContenido(doc)
		if err != nil {
			return err
		}
		ahora := r.now().UTC()
		c.EstadoModeracion = estado
		c.MotivoRechazo = motivo
		c.FechaModeracion = &ahora
		c.ModeradoPor = adminID
		if err := tx.Set(ref, c); err != nil {
			return err
		}
		out = c
		// Después de las lecturas, como pide el outbox.
		return r.events.MaterialDocenteModerado(ctx, tx, *c)
	})
	if err != nil {
		if err == ErrNotFound {
			return nil, err
		}
		return nil, fmt.Errorf("error moderando el contenido: %w", err)
	}
	r.events.Notify()
	return out, nil
}

func (r *teacherContentRepository) Delete(ctx context.Context, id string) error {
	if _, err := r.col().Doc(id).Delete(ctx); err != nil {
		return fmt.Errorf("error borrando el contenido: %w", err)
	}
	return nil
}

func (r *teacherContentRepository) CreateDescarga(ctx context.Context, d *models.DescargaBiblioteca) error {
	ref := r.client.Collection(libraryDownloadsCol).NewDoc()
	if _, err := ref.Create(ctx, d); err != nil {
		return fmt.Errorf("error registrando la descarga: %w", err)
	}
	d.ID = ref.ID
	return nil
}

func (r *teacherContentRepository) ListDescargas(ctx context.Context) ([]models.DescargaBiblioteca, error) {
	iter := r.client.Collection(libraryDownloadsCol).Documents(ctx)
	defer iter.Stop()
	out := make([]models.DescargaBiblioteca, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error listando descargas: %w", err)
		}
		var d models.DescargaBiblioteca
		if err := doc.DataTo(&d); err != nil {
			return nil, fmt.Errorf("error leyendo descarga: %w", err)
		}
		d.ID = doc.Ref.ID
		out = append(out, d)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Fecha.After(out[j].Fecha) })
	return out, nil
}
