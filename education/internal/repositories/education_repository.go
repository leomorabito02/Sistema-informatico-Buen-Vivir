package repositories

import (
	"context"

	"buenvivir/education/internal/models"
)

type EducationRepository interface {
	FindCourses(ctx context.Context) ([]*models.Course, error)
}

type educationRepository struct{}

func NewEducationRepository() EducationRepository {
	return &educationRepository{}
}

func (r *educationRepository) FindCourses(ctx context.Context) ([]*models.Course, error) {
	return []*models.Course{
		{ID: "crs-1", Title: "Capacitación en Oficios Digitales", Category: "Tecnología", DurationHrs: 40},
		{ID: "crs-2", Title: "Taller de Nutrición y Salud Comunitaria", Category: "Salud", DurationHrs: 20},
	}, nil
}
