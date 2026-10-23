package repositories

import (
	"context"
	"fmt"
	"sort"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"buenvivir/campaigns/internal/models"
)

// CampaignsCollection es la colección de Firestore con las campañas.
const CampaignsCollection = "campanas"

// CampaignRepository encapsula el acceso a la colección de campañas.
type CampaignRepository interface {
	Create(ctx context.Context, c models.Campaign) (models.Campaign, error)
	Get(ctx context.Context, id string) (models.Campaign, error)
	// Update reemplaza título, texto, imagen y fecha. No toca el estado
	// ni el autor.
	Update(ctx context.Context, id string, in models.CampaignInput) (models.Campaign, error)
	// SetEstado cambia el estado de la campaña (publicar, HU-49).
	SetEstado(ctx context.Context, id, estado string) (models.Campaign, error)
	// List devuelve las campañas, de la más reciente a la más vieja (por
	// fecha). Con estado vacío devuelve todas.
	List(ctx context.Context, estado string) ([]models.Campaign, error)
}

type campaignRepository struct {
	client *firestore.Client
}

func NewCampaignRepository(client *firestore.Client) CampaignRepository {
	return &campaignRepository{client: client}
}

func (r *campaignRepository) Create(ctx context.Context, c models.Campaign) (models.Campaign, error) {
	ref := r.client.Collection(CampaignsCollection).NewDoc()
	if _, err := ref.Create(ctx, c); err != nil {
		return models.Campaign{}, fmt.Errorf("error creando la campaña: %w", err)
	}
	c.ID = ref.ID
	return c, nil
}

func (r *campaignRepository) Get(ctx context.Context, id string) (models.Campaign, error) {
	doc, err := r.client.Collection(CampaignsCollection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return models.Campaign{}, ErrNotFound
		}
		return models.Campaign{}, fmt.Errorf("error leyendo la campaña: %w", err)
	}
	return campaignFromDoc(doc)
}

func (r *campaignRepository) Update(ctx context.Context, id string, in models.CampaignInput) (models.Campaign, error) {
	return r.update(ctx, id, []firestore.Update{
		{Path: "titulo", Value: in.Titulo},
		{Path: "texto", Value: in.Texto},
		{Path: "imagen_url", Value: in.ImagenURL},
		{Path: "fecha", Value: in.Fecha},
	})
}

func (r *campaignRepository) SetEstado(ctx context.Context, id, estado string) (models.Campaign, error) {
	return r.update(ctx, id, []firestore.Update{{Path: "estado", Value: estado}})
}

// update aplica los cambios y devuelve la campaña como quedó. Update de
// Firestore falla con NotFound si el documento no existe, así que no se
// crean campañas por accidente.
func (r *campaignRepository) update(ctx context.Context, id string, cambios []firestore.Update) (models.Campaign, error) {
	ref := r.client.Collection(CampaignsCollection).Doc(id)
	if _, err := ref.Update(ctx, cambios); err != nil {
		if status.Code(err) == codes.NotFound {
			return models.Campaign{}, ErrNotFound
		}
		return models.Campaign{}, fmt.Errorf("error actualizando la campaña: %w", err)
	}
	return r.Get(ctx, id)
}

func (r *campaignRepository) List(ctx context.Context, estado string) ([]models.Campaign, error) {
	q := r.client.Collection(CampaignsCollection).Query
	if estado != "" {
		q = q.Where("estado", "==", estado)
	}
	docs, err := q.Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("error listando campañas: %w", err)
	}
	campaigns := make([]models.Campaign, 0, len(docs))
	for _, d := range docs {
		c, err := campaignFromDoc(d)
		if err != nil {
			return nil, err
		}
		campaigns = append(campaigns, c)
	}
	// Se ordena acá y no con OrderBy: Where + OrderBy sobre campos
	// distintos necesita un índice compuesto en Firestore.
	sort.SliceStable(campaigns, func(i, j int) bool {
		if campaigns[i].Fecha != campaigns[j].Fecha {
			return campaigns[i].Fecha > campaigns[j].Fecha
		}
		return campaigns[i].ID < campaigns[j].ID
	})
	return campaigns, nil
}

func campaignFromDoc(doc *firestore.DocumentSnapshot) (models.Campaign, error) {
	var c models.Campaign
	if err := doc.DataTo(&c); err != nil {
		return models.Campaign{}, fmt.Errorf("error leyendo la campaña %s: %w", doc.Ref.ID, err)
	}
	c.ID = doc.Ref.ID
	return c, nil
}
