// Package reportes guarda el historial de reportes exportados del
// dashboard (HU-55, tabla ReporteExportado del DER): los datos en
// Firestore y el PDF en Cloud Storage. Es la única persistencia del
// Dashboard; los indicadores se siguen calculando en vivo.
package reportes

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// Colección de Firestore y carpeta del bucket.
const (
	Collection    = "reportes_exportados"
	CarpetaBucket = "reportes/"
	TipoDashboard = "dashboard"
	MaxPDFBytes   = 10 << 20
	maxNombre     = 120
	maxListado    = 100
	fechaLayout   = "2006-01-02"
)

var (
	ErrNoEncontrado = errors.New("el reporte no existe")
	idPattern       = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)
	modulos         = map[string]bool{"todos": true, "usuarios": true, "educacion": true, "promocion": true, "insercion": true}
)

// ErrValidacion: el pedido no cumple las reglas (se responde 400).
type ErrValidacion struct{ Mensaje string }

func (e *ErrValidacion) Error() string { return e.Mensaje }

// Filtros son los filtros que tenía el dashboard al exportar.
type Filtros struct {
	Modulo     string  `json:"modulo" firestore:"modulo"`
	AreaID     *string `json:"areaId" firestore:"area_id"`
	AreaNombre *string `json:"areaNombre" firestore:"area_nombre"`
}

// Reporte es un ReporteExportado del DER. generado_por es el id del
// usuario (FK lógica a Autenticación); generado_por_nombre se guarda
// para mostrar el historial sin consultar a Autenticación.
type Reporte struct {
	ID                string    `json:"id" firestore:"-"`
	TipoReporte       string    `json:"tipo_reporte" firestore:"tipo_reporte"`
	PeriodoDesde      string    `json:"periodo_desde" firestore:"periodo_desde"`
	PeriodoHasta      string    `json:"periodo_hasta" firestore:"periodo_hasta"`
	FiltrosAplicados  Filtros   `json:"filtros_aplicados" firestore:"filtros_aplicados"`
	GeneradoPor       string    `json:"generado_por" firestore:"generado_por"`
	GeneradoPorNombre string    `json:"generado_por_nombre" firestore:"generado_por_nombre"`
	FechaGeneracion   time.Time `json:"fecha_generacion" firestore:"fecha_generacion"`
	URLArchivo        string    `json:"url_archivo" firestore:"url_archivo"`
}

// Metadatos es lo que manda el frontend junto con el PDF.
type Metadatos struct {
	PeriodoDesde      string  `json:"periodo_desde"`
	PeriodoHasta      string  `json:"periodo_hasta"`
	FiltrosAplicados  Filtros `json:"filtros_aplicados"`
	GeneradoPorNombre string  `json:"generado_por_nombre"`
}

// Repository guarda los datos del reporte.
type Repository interface {
	NuevoID() string
	Crear(ctx context.Context, r Reporte) error
	Obtener(ctx context.Context, id string) (Reporte, error)
	// Listar devuelve los últimos `limite` reportes, del más nuevo al más viejo.
	Listar(ctx context.Context, limite int) ([]Reporte, error)
}

// Storage guarda el PDF.
type Storage interface {
	Subir(ctx context.Context, objeto, contentType string, data []byte) error
	Abrir(ctx context.Context, objeto string) (io.ReadCloser, error)
	Borrar(ctx context.Context, objeto string) error
}

type Service struct {
	repo    Repository
	storage Storage
	bucket  string
	ahora   func() time.Time
}

func NewService(repo Repository, storage Storage, bucket string) *Service {
	return &Service{repo: repo, storage: storage, bucket: bucket, ahora: time.Now}
}

// Registrar valida, sube el PDF y guarda el reporte. Si guardar los datos
// falla, borra el PDF para no dejar archivos sueltos en el bucket.
func (s *Service) Registrar(ctx context.Context, meta Metadatos, pdf []byte, usuarioID string) (Reporte, error) {
	if err := validar(&meta, pdf, usuarioID); err != nil {
		return Reporte{}, err
	}
	id := s.repo.NuevoID()
	objeto := CarpetaBucket + id + ".pdf"
	r := Reporte{
		ID:                id,
		TipoReporte:       TipoDashboard,
		PeriodoDesde:      meta.PeriodoDesde,
		PeriodoHasta:      meta.PeriodoHasta,
		FiltrosAplicados:  meta.FiltrosAplicados,
		GeneradoPor:       usuarioID,
		GeneradoPorNombre: meta.GeneradoPorNombre,
		FechaGeneracion:   s.ahora().UTC(),
		URLArchivo:        fmt.Sprintf("gs://%s/%s", s.bucket, objeto),
	}
	if err := s.storage.Subir(ctx, objeto, "application/pdf", pdf); err != nil {
		return Reporte{}, err
	}
	if err := s.repo.Crear(ctx, r); err != nil {
		_ = s.storage.Borrar(ctx, objeto)
		return Reporte{}, err
	}
	return r, nil
}

func (s *Service) Listar(ctx context.Context) ([]Reporte, error) {
	return s.repo.Listar(ctx, maxListado)
}

// AbrirPDF devuelve el PDF del reporte para volver a descargarlo.
func (s *Service) AbrirPDF(ctx context.Context, id string) (Reporte, io.ReadCloser, error) {
	if !idPattern.MatchString(id) {
		return Reporte{}, nil, ErrNoEncontrado
	}
	r, err := s.repo.Obtener(ctx, id)
	if err != nil {
		return Reporte{}, nil, err
	}
	rc, err := s.storage.Abrir(ctx, CarpetaBucket+id+".pdf")
	if err != nil {
		return Reporte{}, nil, err
	}
	return r, rc, nil
}

func validar(meta *Metadatos, pdf []byte, usuarioID string) error {
	if strings.TrimSpace(usuarioID) == "" {
		return &ErrValidacion{"No se pudo identificar a quien exporta el reporte"}
	}
	if len(pdf) == 0 || !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		return &ErrValidacion{"El archivo tiene que ser un PDF"}
	}
	if len(pdf) > MaxPDFBytes {
		return &ErrValidacion{"El PDF no puede superar los 10 MB"}
	}
	desde, err1 := time.Parse(fechaLayout, meta.PeriodoDesde)
	hasta, err2 := time.Parse(fechaLayout, meta.PeriodoHasta)
	if err1 != nil || err2 != nil {
		return &ErrValidacion{"El período tiene que tener fechas AAAA-MM-DD"}
	}
	if hasta.Before(desde) {
		return &ErrValidacion{"La fecha de fin del período no puede ser anterior a la de inicio"}
	}
	if !modulos[meta.FiltrosAplicados.Modulo] {
		return &ErrValidacion{"Módulo de filtro desconocido"}
	}
	meta.GeneradoPorNombre = strings.TrimSpace(meta.GeneradoPorNombre)
	if utf8.RuneCountInString(meta.GeneradoPorNombre) > maxNombre {
		return &ErrValidacion{"El nombre de quien exporta es demasiado largo"}
	}
	for _, v := range []*string{meta.FiltrosAplicados.AreaID, meta.FiltrosAplicados.AreaNombre} {
		if v != nil && utf8.RuneCountInString(*v) > maxNombre {
			return &ErrValidacion{"El filtro de área es demasiado largo"}
		}
	}
	return nil
}
