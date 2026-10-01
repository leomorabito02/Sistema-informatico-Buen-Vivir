package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"buenvivir/human-promotion/internal/crypto"
	"buenvivir/human-promotion/internal/models"
	"buenvivir/human-promotion/internal/repositories"
)

// FichaService implementa la lógica de la ficha de proceso personal
// (HU-29, HU-30, HU-31): validación, sanitización y cifrado de los
// campos confidenciales antes de persistir.
type FichaService interface {
	Create(ctx context.Context, actorID string, input models.CreateFichaInput) (*models.FichaProcesoPersonal, error)
	GetByID(ctx context.Context, id string) (*models.FichaProcesoPersonal, error)
	GetByParticipante(ctx context.Context, participanteID string) (*models.FichaProcesoPersonal, error)
	Update(ctx context.Context, id string, input models.UpdateFichaInput) (*models.FichaProcesoPersonal, error)
	// GetDetalle y GetDetalleByParticipante agregan a la ficha los datos
	// de solo lectura que viven en Auth (HU-31).
	GetDetalle(ctx context.Context, id string) (*models.FichaDetalle, error)
	GetDetalleByParticipante(ctx context.Context, participanteID string) (*models.FichaDetalle, error)
}

type fichaService struct {
	repo   repositories.FichaRepository
	cipher crypto.FieldCipher
	auth   repositories.AuthDirectory
	now    func() time.Time
}

func NewFichaService(repo repositories.FichaRepository, cipher crypto.FieldCipher, auth repositories.AuthDirectory) FichaService {
	return &fichaService{repo: repo, cipher: cipher, auth: auth, now: time.Now}
}

func (s *fichaService) Create(ctx context.Context, actorID string, input models.CreateFichaInput) (*models.FichaProcesoPersonal, error) {
	situacion, avances, err := validateContenido(input.SituacionActual, input.Avances)
	if err != nil {
		return nil, err
	}
	participanteID := input.ParticipanteID
	if participanteID == "" {
		return nil, NewValidationError(MsgCamposObligatorios)
	}
	if !validID(participanteID) {
		return nil, NewValidationError("participante_id inválido")
	}
	// La ficha solo se crea para un participante que existe en Auth.
	if _, err := s.auth.GetParticipante(ctx, participanteID); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, NewValidationError("El participante seleccionado no existe")
		}
		return nil, err
	}

	encSituacion, encAvances, err := s.encryptContenido(situacion, avances)
	if err != nil {
		return nil, err
	}

	updatedAt := s.now().UTC()
	id, err := s.repo.Create(ctx, models.FichaDocument{
		ParticipanteID:  participanteID,
		SituacionActual: encSituacion,
		Avances:         encAvances,
		CreadoPor:       actorID,
		UpdatedAt:       updatedAt,
	})
	if err != nil {
		return nil, err
	}

	return &models.FichaProcesoPersonal{
		ID:              id,
		ParticipanteID:  participanteID,
		SituacionActual: situacion,
		Avances:         avances,
		CreadoPor:       actorID,
		UpdatedAt:       updatedAt,
	}, nil
}

func (s *fichaService) GetByID(ctx context.Context, id string) (*models.FichaProcesoPersonal, error) {
	if !validID(id) {
		return nil, repositories.ErrNotFound
	}
	doc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.decryptFicha(id, doc)
}

func (s *fichaService) GetByParticipante(ctx context.Context, participanteID string) (*models.FichaProcesoPersonal, error) {
	if !validID(participanteID) {
		return nil, NewValidationError("participante_id inválido")
	}
	id, doc, err := s.repo.GetByParticipante(ctx, participanteID)
	if err != nil {
		return nil, err
	}
	return s.decryptFicha(id, doc)
}

func (s *fichaService) Update(ctx context.Context, id string, input models.UpdateFichaInput) (*models.FichaProcesoPersonal, error) {
	if !validID(id) {
		return nil, repositories.ErrNotFound
	}
	situacion, avances, err := validateContenido(input.SituacionActual, input.Avances)
	if err != nil {
		return nil, err
	}

	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	encSituacion, encAvances, err := s.encryptContenido(situacion, avances)
	if err != nil {
		return nil, err
	}

	updatedAt := s.now().UTC()
	if err := s.repo.UpdateContenido(ctx, id, encSituacion, encAvances, updatedAt); err != nil {
		return nil, err
	}

	return &models.FichaProcesoPersonal{
		ID:              id,
		ParticipanteID:  current.ParticipanteID,
		SituacionActual: situacion,
		Avances:         avances,
		CreadoPor:       current.CreadoPor,
		UpdatedAt:       updatedAt,
	}, nil
}

// validateContenido sanitiza y valida los dos campos obligatorios de la
// ficha. Un campo que queda vacío después de sanitizar cuenta como faltante.
func validateContenido(situacionRaw, avancesRaw string) (string, string, error) {
	situacion := sanitizeText(situacionRaw)
	avances := sanitizeText(avancesRaw)
	if situacion == "" || avances == "" {
		return "", "", NewValidationError(MsgCamposObligatorios)
	}
	if tooLong(situacion, maxTextoFicha) || tooLong(avances, maxTextoFicha) {
		return "", "", NewValidationError(fmt.Sprintf("situación actual y avances admiten hasta %d caracteres", maxTextoFicha))
	}
	return situacion, avances, nil
}

func (s *fichaService) encryptContenido(situacion, avances string) (string, string, error) {
	encSituacion, err := s.cipher.Encrypt(situacion)
	if err != nil {
		return "", "", fmt.Errorf("error cifrando la ficha: %w", err)
	}
	encAvances, err := s.cipher.Encrypt(avances)
	if err != nil {
		return "", "", fmt.Errorf("error cifrando la ficha: %w", err)
	}
	return encSituacion, encAvances, nil
}

func (s *fichaService) decryptFicha(id string, doc *models.FichaDocument) (*models.FichaProcesoPersonal, error) {
	situacion, err := s.cipher.Decrypt(doc.SituacionActual)
	if err != nil {
		return nil, fmt.Errorf("error descifrando la ficha %s: %w", id, err)
	}
	avances, err := s.cipher.Decrypt(doc.Avances)
	if err != nil {
		return nil, fmt.Errorf("error descifrando la ficha %s: %w", id, err)
	}
	return &models.FichaProcesoPersonal{
		ID:              id,
		ParticipanteID:  doc.ParticipanteID,
		SituacionActual: situacion,
		Avances:         avances,
		CreadoPor:       doc.CreadoPor,
		UpdatedAt:       doc.UpdatedAt,
	}, nil
}

func (s *fichaService) GetDetalle(ctx context.Context, id string) (*models.FichaDetalle, error) {
	ficha, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.completarDetalle(ctx, ficha), nil
}

func (s *fichaService) GetDetalleByParticipante(ctx context.Context, participanteID string) (*models.FichaDetalle, error) {
	ficha, err := s.GetByParticipante(ctx, participanteID)
	if err != nil {
		return nil, err
	}
	return s.completarDetalle(ctx, ficha), nil
}

// completarDetalle suma los datos de Auth. Si Auth no responde, la ficha
// se devuelve igual (sin esos datos): el contenido confidencial es de
// este servicio y no debe quedar inaccesible por una caída ajena.
func (s *fichaService) completarDetalle(ctx context.Context, ficha *models.FichaProcesoPersonal) *models.FichaDetalle {
	detalle := &models.FichaDetalle{FichaProcesoPersonal: *ficha, Habilidades: []string{}}

	info, err := s.auth.GetParticipante(ctx, ficha.ParticipanteID)
	if err != nil {
		slog.Warn("no se pudieron obtener los datos del participante", slog.String("ficha_id", ficha.ID), slog.String("error", err.Error()))
	} else {
		detalle.Participante = &models.DatosParticipante{
			Nombre:          info.Nombre,
			Apellido:        info.Apellido,
			FechaNacimiento: info.FechaNacimiento,
			Edad:            calcularEdad(info.FechaNacimiento, s.now()),
		}
	}

	if ficha.CreadoPor != "" {
		nombres, err := s.auth.GetNombresUsuarios(ctx, []string{ficha.CreadoPor})
		if err != nil {
			slog.Warn("no se pudo obtener el nombre del creador de la ficha", slog.String("ficha_id", ficha.ID), slog.String("error", err.Error()))
		} else {
			detalle.CreadoPorNombre = nombres[ficha.CreadoPor]
		}
	}
	return detalle
}

// calcularEdad devuelve la edad en años cumplidos a la fecha now, o nil
// si la fecha de nacimiento no está cargada o es inválida.
func calcularEdad(fechaNacimiento string, now time.Time) *int {
	nacimiento, err := time.Parse(models.FechaLayout, fechaNacimiento)
	if err != nil || nacimiento.After(now) {
		return nil
	}
	edad := now.Year() - nacimiento.Year()
	if now.Month() < nacimiento.Month() || (now.Month() == nacimiento.Month() && now.Day() < nacimiento.Day()) {
		edad--
	}
	return &edad
}
