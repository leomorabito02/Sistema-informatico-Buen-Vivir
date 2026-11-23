package services

import (
	"context"
	"errors"
	"log"
	"net"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
)

const (
	// maxSuggestions es cuántas habilidades se le piden a la IA por
	// actividad. Menos que maxSkillsPerActivity, para dejar lugar a las
	// que el docente o el administrador agreguen a mano.
	maxSuggestions = 5
	// maxJustificationLength recorta justificaciones demasiado largas: en
	// pantalla tiene que ser una oración.
	maxJustificationLength = 300
	// retryBackoff es la espera entre reintentos de un mismo proveedor
	// (se multiplica por el número de intento: 0,5 s, 1 s, 1,5 s).
	retryBackoff = 500 * time.Millisecond
)

// MsgSuggestionsUnavailable es el texto que ve quien pidió sugerencias
// cuando ninguna IA pudo responder. Deja claro que puede seguir sin la IA.
const MsgSuggestionsUnavailable = "No pudimos generar sugerencias en este momento. Podés cargarlas a mano o probar de nuevo más tarde."

// SugeridorHabilidades es la interfaz común de los proveedores de IA (Gemini y
// OpenAI). Permite cambiarlos sin tocar la lógica y probarla con mocks.
type SugeridorHabilidades interface {
	Nombre() string
	Sugerir(ctx context.Context, p models.SkillPrompt) ([]models.SkillSuggestion, error)
}

// SkillSuggestionService pide sugerencias de habilidades para una
// actividad (HU-21 / HU-58). Las sugerencias NO se guardan: viven en el
// formulario hasta que alguien las acepta y guarda la actividad.
type SkillSuggestionService interface {
	// Suggest arma el pedido, prueba los proveedores en orden y devuelve
	// las sugerencias ya filtradas. course puede ser nil (por ejemplo, un
	// curso en borrador que la Content Delivery API todavía no ve): en
	// ese caso la IA trabaja solo con los datos de la actividad.
	Suggest(ctx context.Context, req models.SkillSuggestionRequest, course *models.Course) (*models.SkillSuggestionResponse, error)
}

type skillSuggestionService struct {
	repo       repositories.SkillRepository
	providers  []SugeridorHabilidades
	timeout    time.Duration
	maxRetries int
	sleep      func(ctx context.Context, d time.Duration) error
}

// NewSkillSuggestionService recibe los proveedores en orden de
// preferencia (primero Gemini, después OpenAI). Puede recibir una lista
// vacía si no hay ninguno configurado: el servicio arranca igual y cada
// pedido responde MsgSuggestionsUnavailable.
func NewSkillSuggestionService(repo repositories.SkillRepository, providers []SugeridorHabilidades, timeout time.Duration, maxRetries int) SkillSuggestionService {
	return &skillSuggestionService{repo: repo, providers: providers, timeout: timeout, maxRetries: maxRetries, sleep: sleepCtx}
}

func (s *skillSuggestionService) Suggest(ctx context.Context, req models.SkillSuggestionRequest, course *models.Course) (*models.SkillSuggestionResponse, error) {
	req.Titulo = strings.TrimSpace(req.Titulo)
	req.Consigna = strings.TrimSpace(req.Consigna)
	if req.Titulo == "" || req.Consigna == "" {
		return nil, NewValidationError("completá el título y la consigna de la actividad para pedir sugerencias")
	}

	catalog, err := s.repo.ListCatalog(ctx)
	if err != nil {
		return nil, err
	}
	sortSkills(catalog)
	catalogByKey := make(map[string]models.Skill, len(catalog))
	catalogNames := make([]string, len(catalog))
	for i, sk := range catalog {
		catalogByKey[sk.ID] = sk
		catalogNames[i] = sk.Nombre
	}

	excluded, excludedNames, err := s.excludedSkills(ctx, req, catalogByKey)
	if err != nil {
		return nil, err
	}

	prompt := models.SkillPrompt{
		ActividadTitulo: req.Titulo,
		Consigna:        req.Consigna,
		Catalogo:        catalogNames,
		YaAsignadas:     excludedNames,
		MaxSugerencias:  maxSuggestions,
	}
	if course != nil {
		prompt.CursoTitulo = course.Title
		if course.TrainingArea != nil {
			prompt.AreaFormativa = course.TrainingArea.Name
		}
	}

	for _, provider := range s.providers {
		raw, err := s.callWithRetries(ctx, provider, prompt)
		if err != nil {
			if ctx.Err() != nil {
				// Quien pidió se fue (cerró la pestaña, se cortó la
				// conexión): no tiene sentido seguir gastando en el respaldo.
				return nil, ctx.Err()
			}
			log.Printf("sugerencia de habilidades: %s falló, se prueba el siguiente proveedor: %v", provider.Nombre(), err)
			continue
		}
		valid, ok := validateSuggestions(raw)
		if !ok {
			log.Printf("sugerencia de habilidades: %s devolvió sugerencias inválidas, se prueba el siguiente proveedor", provider.Nombre())
			continue
		}
		// Auditoría (pendiente, HU-35): acá va el registro en LogAuditoria
		// del pedido de sugerencias y del proveedor que respondió, cuando
		// auth-service exponga cómo escribirlo (ver docs/ia-habilidades.md).
		log.Printf("sugerencia de habilidades: respondió %s (%d sugerencias)", provider.Nombre(), len(valid))
		return &models.SkillSuggestionResponse{
			Proveedor:   provider.Nombre(),
			Sugerencias: filterSuggestions(valid, excluded, catalogByKey),
		}, nil
	}

	if len(s.providers) == 0 {
		log.Printf("sugerencia de habilidades: no hay proveedores de IA configurados (VERTEX_PROJECT_ID / OPENAI_API_KEY)")
	}
	return nil, NewUnavailableError(MsgSuggestionsUnavailable)
}

// excludedSkills junta las habilidades que no hay que volver a sugerir:
// las que la actividad ya tiene guardadas y las que ya están en el
// formulario sin guardar. Devuelve sus claves normalizadas (para
// filtrar) y sus nombres (para avisarle a la IA).
func (s *skillSuggestionService) excludedSkills(ctx context.Context, req models.SkillSuggestionRequest, catalogByKey map[string]models.Skill) (map[string]bool, []string, error) {
	excluded := make(map[string]bool)
	var names []string
	add := func(name string) {
		key := models.SkillKey(name)
		if key == "" || excluded[key] {
			return
		}
		excluded[key] = true
		names = append(names, models.CleanSkillName(name))
	}

	if strings.TrimSpace(req.ActividadID) != "" {
		links, err := s.repo.ListLinksByActivity(ctx, req.ActividadID)
		if err != nil {
			return nil, nil, err
		}
		for _, l := range links {
			if sk, ok := catalogByKey[l.HabilidadID]; ok {
				add(sk.Nombre)
			}
		}
	}
	for _, n := range req.Excluir {
		add(n)
	}
	sort.Strings(names)
	return excluded, names, nil
}

// callWithRetries llama a un proveedor con un tiempo máximo por intento
// y reintenta solo las fallas pasajeras. Una respuesta con formato
// inválido no se reintenta: se pasa directo al siguiente proveedor.
func (s *skillSuggestionService) callWithRetries(ctx context.Context, provider SugeridorHabilidades, prompt models.SkillPrompt) ([]models.SkillSuggestion, error) {
	var lastErr error
	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		if attempt > 0 {
			if err := s.sleep(ctx, time.Duration(attempt)*retryBackoff); err != nil {
				return nil, err
			}
		}
		attemptCtx, cancel := context.WithTimeout(ctx, s.timeout)
		result, err := provider.Sugerir(attemptCtx, prompt)
		cancel()
		if err == nil {
			return result, nil
		}
		lastErr = err
		if ctx.Err() != nil || !isRetryableAIError(err) {
			break
		}
	}
	return nil, lastErr
}

// isRetryableAIError decide si vale la pena reintentar el mismo
// proveedor: se agotó el tiempo, falló la red, o el proveedor respondió
// "demasiados pedidos" / error interno.
func isRetryableAIError(err error) bool {
	if errors.Is(err, repositories.ErrAIInvalidResponse) {
		return false
	}
	var httpErr *repositories.AIHTTPError
	if errors.As(err, &httpErr) {
		return httpErr.Retryable()
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr)
}

// validateSuggestions descarta las sugerencias que no cumplen las reglas del catálogo.
// Si ninguna sirve, devuelve ok = false para probar el respaldo. Una lista vacía es válida.
func validateSuggestions(raw []models.SkillSuggestion) ([]models.SkillSuggestion, bool) {
	valid := make([]models.SkillSuggestion, 0, len(raw))
	for _, sg := range raw {
		nombre := models.CleanSkillName(sg.Nombre)
		if models.SkillKey(nombre) == "" || utf8.RuneCountInString(nombre) > maxSkillNameLength || !sg.Categoria.IsValid() {
			continue
		}
		just := strings.Join(strings.Fields(sg.Justificacion), " ")
		if utf8.RuneCountInString(just) > maxJustificationLength {
			just = string([]rune(just)[:maxJustificationLength]) + "…"
		}
		valid = append(valid, models.SkillSuggestion{Nombre: nombre, Categoria: sg.Categoria, Justificacion: just})
	}
	if len(raw) > 0 && len(valid) == 0 {
		return nil, false
	}
	return valid, true
}

// filterSuggestions descarta las que la actividad ya tiene y las repetidas, y reemplaza
// las que ya existen en el catálogo por la versión del catálogo (con su id).
func filterSuggestions(valid []models.SkillSuggestion, excluded map[string]bool, catalogByKey map[string]models.Skill) []models.SkillSuggestion {
	result := make([]models.SkillSuggestion, 0, len(valid))
	seen := make(map[string]bool, len(valid))
	for _, sg := range valid {
		key := models.SkillKey(sg.Nombre)
		if excluded[key] || seen[key] {
			continue
		}
		seen[key] = true
		if existing, ok := catalogByKey[key]; ok {
			sg.ID = existing.ID
			sg.Nombre = existing.Nombre
			sg.Categoria = existing.Categoria
			sg.Existente = true
		}
		result = append(result, sg)
		if len(result) == maxSuggestions {
			break
		}
	}
	return result
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
