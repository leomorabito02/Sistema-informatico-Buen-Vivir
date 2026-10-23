package services_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
	"buenvivir/education/internal/services"
)

// fakeProvider simula una IA: devuelve, en orden, una respuesta por
// llamada. Nunca hace un pedido real a Gemini ni a OpenAI.
type fakeProvider struct {
	name      string
	responses []fakeResponse
	calls     int
	prompts   []models.SkillPrompt
}

type fakeResponse struct {
	skills []models.SkillSuggestion
	err    error
}

func (f *fakeProvider) Nombre() string { return f.name }

func (f *fakeProvider) Sugerir(_ context.Context, p models.SkillPrompt) ([]models.SkillSuggestion, error) {
	f.prompts = append(f.prompts, p)
	r := f.responses[min(f.calls, len(f.responses)-1)]
	f.calls++
	return r.skills, r.err
}

func ok(skills ...models.SkillSuggestion) fakeResponse { return fakeResponse{skills: skills} }
func fail(err error) fakeResponse                      { return fakeResponse{err: err} }

func sug(nombre string, cat models.SkillCategory) models.SkillSuggestion {
	return models.SkillSuggestion{Nombre: nombre, Categoria: cat, Justificacion: "La actividad lo trabaja."}
}

var (
	errServer   = &repositories.AIHTTPError{Provider: "x", StatusCode: http.StatusInternalServerError}
	errBadKey   = &repositories.AIHTTPError{Provider: "x", StatusCode: http.StatusUnauthorized}
	errBadShape = repositories.ErrAIInvalidResponse
)

func newSuggestionService(repo *MockSkillRepository, providers ...services.SugeridorHabilidades) services.SkillSuggestionService {
	return services.NewSkillSuggestionService(repo, providers, time.Second, 1)
}

func baseRepo(catalog ...models.Skill) *MockSkillRepository {
	repo := new(MockSkillRepository)
	repo.On("ListCatalog", mock.Anything).Return(catalog, nil)
	return repo
}

var validReq = models.SkillSuggestionRequest{CursoID: "c1", Titulo: "Armado de canteros", Consigna: "Armen un cantero en grupo y registren qué plantaron."}

func TestSuggest_GeminiRespondsCorrectly(t *testing.T) {
	gemini := &fakeProvider{name: "gemini", responses: []fakeResponse{ok(sug("Armado de huerta", models.SkillCategoryNaturaleza), sug("Trabajo en equipo", models.SkillCategorySocioemoc))}}
	openai := &fakeProvider{name: "openai", responses: []fakeResponse{ok()}}
	svc := newSuggestionService(baseRepo(), gemini, openai)

	course := &models.Course{Title: "Huerta comunitaria", TrainingArea: &models.TrainingArea{Name: "Ambiente"}}
	resp, err := svc.Suggest(context.Background(), validReq, course)
	require.NoError(t, err)
	assert.Equal(t, "gemini", resp.Proveedor)
	assert.Len(t, resp.Sugerencias, 2)
	assert.Equal(t, 0, openai.calls, "si Gemini responde, no se usa el respaldo")

	// Lo que se le manda a la IA: solo datos de la actividad y el curso.
	p := gemini.prompts[0]
	assert.Equal(t, "Huerta comunitaria", p.CursoTitulo)
	assert.Equal(t, "Ambiente", p.AreaFormativa)
	assert.Equal(t, validReq.Titulo, p.ActividadTitulo)
	assert.Equal(t, validReq.Consigna, p.Consigna)
}

func TestSuggest_GeminiFailsAndOpenAIResponds(t *testing.T) {
	gemini := &fakeProvider{name: "gemini", responses: []fakeResponse{fail(errBadKey)}}
	openai := &fakeProvider{name: "openai", responses: []fakeResponse{ok(sug("Uso de Excel", models.SkillCategoryTecnologia))}}
	svc := newSuggestionService(baseRepo(), gemini, openai)

	resp, err := svc.Suggest(context.Background(), validReq, nil)
	require.NoError(t, err)
	assert.Equal(t, "openai", resp.Proveedor)
	assert.Equal(t, 1, gemini.calls, "un 401 no se reintenta")
}

func TestSuggest_BothFail(t *testing.T) {
	gemini := &fakeProvider{name: "gemini", responses: []fakeResponse{fail(errServer)}}
	openai := &fakeProvider{name: "openai", responses: []fakeResponse{fail(errors.New("connection refused"))}}
	svc := newSuggestionService(baseRepo(), gemini, openai)

	_, err := svc.Suggest(context.Background(), validReq, nil)
	require.Error(t, err)
	assert.True(t, services.IsUnavailableError(err))
	assert.Equal(t, services.MsgSuggestionsUnavailable, err.Error())
	assert.Equal(t, 2, gemini.calls, "un 500 se reintenta una vez")
}

func TestSuggest_NoProvidersConfigured(t *testing.T) {
	svc := newSuggestionService(baseRepo())
	_, err := svc.Suggest(context.Background(), validReq, nil)
	assert.True(t, services.IsUnavailableError(err))
}

func TestSuggest_InvalidFormatGoesStraightToFallback(t *testing.T) {
	gemini := &fakeProvider{name: "gemini", responses: []fakeResponse{fail(errBadShape)}}
	openai := &fakeProvider{name: "openai", responses: []fakeResponse{ok(sug("Uso de Excel", models.SkillCategoryTecnologia))}}
	svc := newSuggestionService(baseRepo(), gemini, openai)

	resp, err := svc.Suggest(context.Background(), validReq, nil)
	require.NoError(t, err)
	assert.Equal(t, "openai", resp.Proveedor)
	assert.Equal(t, 1, gemini.calls, "una respuesta con formato inválido no se reintenta")
}

func TestSuggest_AllSuggestionsInvalidIsTreatedAsInvalidResponse(t *testing.T) {
	// JSON bien formado pero con categorías que no existen y nombres vacíos.
	gemini := &fakeProvider{name: "gemini", responses: []fakeResponse{ok(
		sug("Soldadura", "Técnica"),
		sug("   ", models.SkillCategoryOficio),
	)}}
	openai := &fakeProvider{name: "openai", responses: []fakeResponse{ok(sug("Soldadura", models.SkillCategoryOficio))}}
	svc := newSuggestionService(baseRepo(), gemini, openai)

	resp, err := svc.Suggest(context.Background(), validReq, nil)
	require.NoError(t, err)
	assert.Equal(t, "openai", resp.Proveedor)
	assert.Equal(t, "Soldadura", resp.Sugerencias[0].Nombre)
}

func TestSuggest_EmptyListIsAValidAnswer(t *testing.T) {
	gemini := &fakeProvider{name: "gemini", responses: []fakeResponse{ok()}}
	svc := newSuggestionService(baseRepo(), gemini)

	resp, err := svc.Suggest(context.Background(), validReq, nil)
	require.NoError(t, err)
	assert.Empty(t, resp.Sugerencias)
}

func TestSuggest_RetryThenSuccess(t *testing.T) {
	gemini := &fakeProvider{name: "gemini", responses: []fakeResponse{fail(context.DeadlineExceeded), ok(sug("Uso de Excel", models.SkillCategoryTecnologia))}}
	svc := newSuggestionService(baseRepo(), gemini)

	resp, err := svc.Suggest(context.Background(), validReq, nil)
	require.NoError(t, err)
	assert.Equal(t, "gemini", resp.Proveedor)
	assert.Equal(t, 2, gemini.calls)
}

func TestSuggest_FiltersDuplicatesAgainstActivityFormAndCatalog(t *testing.T) {
	huerta := models.Skill{ID: "armado-de-huerta", Nombre: "Armado de huerta", Categoria: models.SkillCategoryNaturaleza}
	equipo := models.Skill{ID: "trabajo-en-equipo", Nombre: "Trabajo en equipo", Categoria: models.SkillCategorySocioemoc}
	repo := baseRepo(huerta, equipo)
	// La actividad ya tiene "Armado de huerta" guardada.
	repo.On("ListLinksByActivity", mock.Anything, "a1").Return([]models.ActivitySkill{{ActividadID: "a1", HabilidadID: huerta.ID}}, nil)

	gemini := &fakeProvider{name: "gemini", responses: []fakeResponse{ok(
		sug("ARMADO DE HUERTA", models.SkillCategoryNaturaleza), // ya guardada en la actividad
		sug("Riego eficiente", models.SkillCategoryNaturaleza),  // ya está en el formulario
		sug("trabajo en  Equipo", models.SkillCategoryOficio),   // existe en el catálogo
		sug("Compostaje", models.SkillCategoryNaturaleza),       // nueva
		sug("compostaje", models.SkillCategoryNaturaleza),       // repetida en la respuesta
	)}}
	svc := newSuggestionService(repo, gemini)

	req := validReq
	req.ActividadID = "a1"
	req.Excluir = []string{"riego  eficiente"}
	resp, err := svc.Suggest(context.Background(), req, nil)
	require.NoError(t, err)

	require.Len(t, resp.Sugerencias, 2)
	assert.Equal(t, models.SkillSuggestion{ID: "trabajo-en-equipo", Nombre: "Trabajo en equipo", Categoria: models.SkillCategorySocioemoc, Justificacion: "La actividad lo trabaja.", Existente: true}, resp.Sugerencias[0],
		"se reusa la del catálogo, con su nombre y categoría")
	assert.Equal(t, "Compostaje", resp.Sugerencias[1].Nombre)
	assert.False(t, resp.Sugerencias[1].Existente)

	// La IA recibe el catálogo y lo que ya está asignado.
	p := gemini.prompts[0]
	assert.Equal(t, []string{"Armado de huerta", "Trabajo en equipo"}, p.Catalogo)
	assert.Equal(t, []string{"Armado de huerta", "Riego eficiente"}, p.YaAsignadas)
}

func TestSuggest_TruncatesLongJustification(t *testing.T) {
	long := make([]rune, 400)
	for i := range long {
		long[i] = 'a'
	}
	s := sug("Uso de Excel", models.SkillCategoryTecnologia)
	s.Justificacion = string(long)
	gemini := &fakeProvider{name: "gemini", responses: []fakeResponse{ok(s)}}

	resp, err := newSuggestionService(baseRepo(), gemini).Suggest(context.Background(), validReq, nil)
	require.NoError(t, err)
	assert.Equal(t, 301, len([]rune(resp.Sugerencias[0].Justificacion)))
}

func TestSuggest_Validation(t *testing.T) {
	svc := newSuggestionService(baseRepo(), &fakeProvider{name: "gemini", responses: []fakeResponse{ok()}})
	_, err := svc.Suggest(context.Background(), models.SkillSuggestionRequest{CursoID: "c1", Titulo: "x"}, nil)
	assert.True(t, services.IsValidationError(err))
}

func TestSuggest_RepoErrors(t *testing.T) {
	repo := new(MockSkillRepository)
	repo.On("ListCatalog", mock.Anything).Return(nil, errors.New("caído"))
	_, err := newSuggestionService(repo).Suggest(context.Background(), validReq, nil)
	assert.Error(t, err)

	repo2 := baseRepo()
	repo2.On("ListLinksByActivity", mock.Anything, "a1").Return(nil, errors.New("caído"))
	req := validReq
	req.ActividadID = "a1"
	_, err = newSuggestionService(repo2).Suggest(context.Background(), req, nil)
	assert.Error(t, err)
}

func TestSuggest_CallerGoneStopsWithoutFallback(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	gemini := &fakeProvider{name: "gemini", responses: []fakeResponse{{err: errServer}}}
	openai := &fakeProvider{name: "openai", responses: []fakeResponse{ok()}}
	cancel()

	_, err := newSuggestionService(baseRepo(), gemini, openai).Suggest(ctx, validReq, nil)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 0, openai.calls)
}

type cancelProvider struct {
	cancel func()
	err    error
}

func (c *cancelProvider) Nombre() string { return "canceler" }
func (c *cancelProvider) Sugerir(_ context.Context, _ models.SkillPrompt) ([]models.SkillSuggestion, error) {
	c.cancel()
	return nil, c.err
}

func TestSuggest_SleepInterruptedByContextCancel(t *testing.T) {
	// Arrange
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p := &cancelProvider{
		cancel: cancel,
		err:    errServer,
	}
	svc := newSuggestionService(baseRepo(), p)

	// Act
	resp, err := svc.Suggest(ctx, validReq, nil)

	// Assert
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, context.Canceled)
}
