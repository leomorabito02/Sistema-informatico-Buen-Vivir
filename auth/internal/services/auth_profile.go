package services

import (
	"context"
	"errors"
	"fmt"
)

// UpdateMyProfile dynamically updates the specific role table for the user
func (s *authService) UpdateMyProfile(ctx context.Context, roleID, userID string, data map[string]interface{}) error {
	var tableName string
	switch roleID {
	case "role_participant":
		tableName = "participants"
	case "role_company":
		tableName = "companies"
	case "role_teacher":
		tableName = "teachers"
	case "role_volunteer":
		tableName = "volunteers"
	default:
		return errors.New("rol no soporta perfil extra o es inválido")
	}

	// Remove unupdatable fields just in case
	delete(data, "user_id")
	delete(data, "approval_status") // security: they shouldn't auto-approve themselves
	delete(data, "approval_date")
	delete(data, "approved_by")
	delete(data, "created_at")
	delete(data, "updated_at")
	delete(data, "participant_id")
	delete(data, "company_id")
	delete(data, "teacher_id")
	delete(data, "volunteer_id")

	if len(data) == 0 {
		return nil // nothing to update
	}

	err := s.db.WithContext(ctx).Table(tableName).Where("user_id = ?", userID).Updates(data).Error
	if err != nil {
		return fmt.Errorf("error actualizando perfil: %w", err)
	}
	return nil
}
