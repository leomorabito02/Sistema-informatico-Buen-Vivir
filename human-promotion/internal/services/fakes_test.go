package services

import (
	"context"
	"fmt"
	"time"

	"buenvivir/human-promotion/internal/models"
	"buenvivir/human-promotion/internal/repositories"
)

// Repositorios en memoria para probar la lógica de los servicios sin
// Firestore. Guardan lo que les llega (ya cifrado por el servicio).

type memFichas struct {
	docs map[string]models.FichaDocument
	next int
}

func newMemFichas() *memFichas { return &memFichas{docs: map[string]models.FichaDocument{}} }

func (m *memFichas) Create(_ context.Context, doc models.FichaDocument) (string, error) {
	for _, d := range m.docs {
		if d.ParticipanteID == doc.ParticipanteID {
			return "", repositories.ErrAlreadyExists
		}
	}
	m.next++
	id := fmt.Sprintf("f%d", m.next)
	m.docs[id] = doc
	return id, nil
}

func (m *memFichas) GetByID(_ context.Context, id string) (*models.FichaDocument, error) {
	doc, ok := m.docs[id]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	return &doc, nil
}

func (m *memFichas) GetByParticipante(_ context.Context, participanteID string) (string, *models.FichaDocument, error) {
	for id, d := range m.docs {
		if d.ParticipanteID == participanteID {
			return id, &d, nil
		}
	}
	return "", nil, repositories.ErrNotFound
}

func (m *memFichas) UpdateContenido(_ context.Context, id, situacion, avances string, updatedAt time.Time) error {
	doc, ok := m.docs[id]
	if !ok {
		return repositories.ErrNotFound
	}
	doc.SituacionActual, doc.Avances, doc.UpdatedAt = situacion, avances, updatedAt
	m.docs[id] = doc
	return nil
}

func (m *memFichas) FindIDsByParticipantes(_ context.Context, ids []string) (map[string]string, error) {
	out := map[string]string{}
	for id, d := range m.docs {
		for _, p := range ids {
			if d.ParticipanteID == p {
				out[p] = id
			}
		}
	}
	return out, nil
}

func (m *memFichas) DeactivateByParticipante(context.Context, string) (bool, error) {
	return false, nil
}

type memHitos struct {
	docs map[string]models.HitoDocument
	next int
}

func newMemHitos() *memHitos { return &memHitos{docs: map[string]models.HitoDocument{}} }

func (m *memHitos) Create(_ context.Context, doc models.HitoDocument) (string, error) {
	m.next++
	id := fmt.Sprintf("h%d", m.next)
	m.docs[id] = doc
	return id, nil
}

func (m *memHitos) GetByID(_ context.Context, id string) (*models.HitoDocument, error) {
	doc, ok := m.docs[id]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	return &doc, nil
}

func (m *memHitos) Update(_ context.Context, id string, doc models.HitoDocument) error {
	if _, ok := m.docs[id]; !ok {
		return repositories.ErrNotFound
	}
	m.docs[id] = doc
	return nil
}

func (m *memHitos) ListByFicha(_ context.Context, fichaID string) (map[string]models.HitoDocument, error) {
	out := map[string]models.HitoDocument{}
	for id, d := range m.docs {
		if d.FichaID == fichaID {
			out[id] = d
		}
	}
	return out, nil
}

// authConNombres simula auth-service con participantes y usuarios conocidos.
type authConNombres struct {
	fakeAuth
	participantes map[string]models.ParticipanteInfo
	usuarios      map[string]string
	err           error
}

func (a authConNombres) GetParticipante(_ context.Context, id string) (*models.ParticipanteInfo, error) {
	if a.err != nil {
		return nil, a.err
	}
	p, ok := a.participantes[id]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	return &p, nil
}

func (a authConNombres) GetNombresUsuarios(_ context.Context, ids []string) (map[string]string, error) {
	if a.err != nil {
		return nil, a.err
	}
	out := map[string]string{}
	for _, id := range ids {
		if n, ok := a.usuarios[id]; ok {
			out[id] = n
		}
	}
	return out, nil
}
