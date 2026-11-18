package services

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"buenvivir/human-promotion/internal/crypto"
	"buenvivir/human-promotion/internal/models"
	"buenvivir/human-promotion/internal/repositories"
)

// HitoService implementa el alta, la edición y la línea de tiempo de
// los hitos de una ficha (HU-32, HU-33).
type HitoService interface {
	Create(ctx context.Context, actorID, fichaID string, input models.HitoInput) (*models.Hito, error)
	Update(ctx context.Context, fichaID, hitoID string, input models.HitoInput) (*models.Hito, error)
	// ListByFicha devuelve todos los hitos de la ficha en orden
	// cronológico (fecha ascendente).
	ListByFicha(ctx context.Context, fichaID string) ([]models.Hito, error)
	// ListDetalle es la línea de tiempo profesional (HU-33): ListByFicha
	// más el nombre de cada responsable.
	ListDetalle(ctx context.Context, fichaID string) ([]models.HitoDetalle, error)
}

type hitoService struct {
	hitos  repositories.HitoRepository
	fichas repositories.FichaRepository
	cipher crypto.FieldCipher
	auth   repositories.AuthDirectory
}

func NewHitoService(hitos repositories.HitoRepository, fichas repositories.FichaRepository, cipher crypto.FieldCipher, auth repositories.AuthDirectory) HitoService {
	return &hitoService{hitos: hitos, fichas: fichas, cipher: cipher, auth: auth}
}

func (s *hitoService) Create(ctx context.Context, actorID, fichaID string, input models.HitoInput) (*models.Hito, error) {
	if err := s.ensureFicha(ctx, fichaID); err != nil {
		return nil, err
	}

	fecha, descripcion, err := validateHito(input)
	if err != nil {
		return nil, err
	}
	responsable := input.ResponsableID
	if responsable == "" {
		responsable = actorID
	}
	if !validID(responsable) {
		return nil, NewValidationError("responsable_id inválido")
	}
	if responsable != actorID {
		if err := s.ensureUsuario(ctx, responsable); err != nil {
			return nil, err
		}
	}
	visible := input.VisibleParticipante != nil && *input.VisibleParticipante

	encDescripcion, err := s.cipher.Encrypt(descripcion)
	if err != nil {
		return nil, fmt.Errorf("error cifrando el hito: %w", err)
	}

	id, err := s.hitos.Create(ctx, models.HitoDocument{
		FichaID:             fichaID,
		Fecha:               fecha,
		Descripcion:         encDescripcion,
		ResponsableID:       responsable,
		VisibleParticipante: visible,
	})
	if err != nil {
		return nil, err
	}

	return &models.Hito{
		ID:                  id,
		FichaID:             fichaID,
		Fecha:               fecha,
		Descripcion:         descripcion,
		ResponsableID:       responsable,
		VisibleParticipante: visible,
	}, nil
}

func (s *hitoService) Update(ctx context.Context, fichaID, hitoID string, input models.HitoInput) (*models.Hito, error) {
	if !validID(fichaID) || !validID(hitoID) {
		return nil, repositories.ErrNotFound
	}
	current, err := s.hitos.GetByID(ctx, hitoID)
	if err != nil {
		return nil, err
	}
	// Un hito solo se edita a través de su propia ficha.
	if current.FichaID != fichaID {
		return nil, repositories.ErrNotFound
	}

	fecha, descripcion, err := validateHito(input)
	if err != nil {
		return nil, err
	}
	responsable := current.ResponsableID
	if input.ResponsableID != "" && input.ResponsableID != current.ResponsableID {
		if !validID(input.ResponsableID) {
			return nil, NewValidationError("responsable_id inválido")
		}
		if err := s.ensureUsuario(ctx, input.ResponsableID); err != nil {
			return nil, err
		}
		responsable = input.ResponsableID
	}
	visible := current.VisibleParticipante
	if input.VisibleParticipante != nil {
		visible = *input.VisibleParticipante
	}

	encDescripcion, err := s.cipher.Encrypt(descripcion)
	if err != nil {
		return nil, fmt.Errorf("error cifrando el hito: %w", err)
	}

	if err := s.hitos.Update(ctx, hitoID, models.HitoDocument{
		FichaID:             fichaID,
		Fecha:               fecha,
		Descripcion:         encDescripcion,
		ResponsableID:       responsable,
		VisibleParticipante: visible,
	}); err != nil {
		return nil, err
	}

	return &models.Hito{
		ID:                  hitoID,
		FichaID:             fichaID,
		Fecha:               fecha,
		Descripcion:         descripcion,
		ResponsableID:       responsable,
		VisibleParticipante: visible,
	}, nil
}

func (s *hitoService) ListByFicha(ctx context.Context, fichaID string) ([]models.Hito, error) {
	if err := s.ensureFicha(ctx, fichaID); err != nil {
		return nil, err
	}

	docs, err := s.hitos.ListByFicha(ctx, fichaID)
	if err != nil {
		return nil, err
	}

	out := make([]models.Hito, 0, len(docs))
	for id, doc := range docs {
		descripcion, err := s.cipher.Decrypt(doc.Descripcion)
		if err != nil {
			return nil, fmt.Errorf("error descifrando el hito %s: %w", id, err)
		}
		out = append(out, models.Hito{
			ID:                  id,
			FichaID:             doc.FichaID,
			Fecha:               doc.Fecha,
			Descripcion:         descripcion,
			ResponsableID:       doc.ResponsableID,
			VisibleParticipante: doc.VisibleParticipante,
		})
	}

	// Fecha AAAA-MM-DD ordena bien como texto; el id desempata hitos del
	// mismo día para que el orden sea estable entre consultas.
	sort.Slice(out, func(i, j int) bool {
		if out[i].Fecha != out[j].Fecha {
			return out[i].Fecha < out[j].Fecha
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func (s *hitoService) ensureFicha(ctx context.Context, fichaID string) error {
	if !validID(fichaID) {
		return repositories.ErrNotFound
	}
	_, err := s.fichas.GetByID(ctx, fichaID)
	return err
}

func (s *hitoService) ListDetalle(ctx context.Context, fichaID string) ([]models.HitoDetalle, error) {
	hitos, err := s.ListByFicha(ctx, fichaID)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(hitos))
	seen := make(map[string]bool, len(hitos))
	for _, h := range hitos {
		if !seen[h.ResponsableID] {
			seen[h.ResponsableID] = true
			ids = append(ids, h.ResponsableID)
		}
	}
	// Si Auth no responde, la línea de tiempo se muestra igual, sin nombres.
	nombres, err := s.auth.GetNombresUsuarios(ctx, ids)
	if err != nil {
		slog.Warn("no se pudieron obtener los nombres de los responsables", slog.String("ficha_id", fichaID), slog.String("error", err.Error()))
		nombres = map[string]string{}
	}

	out := make([]models.HitoDetalle, 0, len(hitos))
	for _, h := range hitos {
		out = append(out, models.HitoDetalle{Hito: h, ResponsableNombre: nombres[h.ResponsableID]})
	}
	return out, nil
}

// ensureUsuario verifica en Auth que el responsable elegido exista.
func (s *hitoService) ensureUsuario(ctx context.Context, userID string) error {
	nombres, err := s.auth.GetNombresUsuarios(ctx, []string{userID})
	if err != nil {
		return err
	}
	if _, ok := nombres[userID]; !ok {
		return NewValidationError("El responsable seleccionado no existe")
	}
	return nil
}

func validateHito(input models.HitoInput) (string, string, error) {
	descripcion := sanitizeText(input.Descripcion)
	if input.Fecha == "" || descripcion == "" {
		return "", "", NewValidationError(MsgCamposObligatorios)
	}
	fecha, ok := parseFechaHito(input.Fecha)
	if !ok {
		return "", "", NewValidationError("la fecha del hito debe tener formato AAAA-MM-DD")
	}
	if tooLong(descripcion, maxDescripcionHito) {
		return "", "", NewValidationError(fmt.Sprintf("la descripción admite hasta %d caracteres", maxDescripcionHito))
	}
	return fecha, descripcion, nil
}
