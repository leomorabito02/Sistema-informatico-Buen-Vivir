package dto

type CreateJobApplicationRequest struct {
	OfferID       string `json:"oferta_id" validate:"required"`
	ParticipantID string `json:"participante_id,omitempty"`
}

type UpdateJobApplicationStatusRequest struct {
	Status string `json:"estado" validate:"required"`
}
