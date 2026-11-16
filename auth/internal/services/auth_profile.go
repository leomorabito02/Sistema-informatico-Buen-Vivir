package services

import (
	"context"
	"errors"
	"fmt"
)

// UpdateMyProfile dynamically updates the specific role table for the user
func (s *authService) UpdateMyProfile(ctx context.Context, roleID, userID string, data map[string]interface{}) error {
	var updateBaseUser bool
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("error obteniendo usuario: %w", err)
	}

	if email, ok := data["email"].(string); ok {
		user.Email = email
		updateBaseUser = true
		delete(data, "email")
	}
	if phone, ok := data["phone"].(string); ok {
		user.Phone = phone
		updateBaseUser = true
		delete(data, "phone")
	}
	if firstName, ok := data["first_name"].(string); ok && firstName != "" {
		user.FirstName = firstName
		updateBaseUser = true
		delete(data, "first_name")
	}
	if lastName, ok := data["last_name"].(string); ok && lastName != "" {
		user.LastName = lastName
		updateBaseUser = true
		delete(data, "last_name")
	}

	if updateBaseUser {
		if err := s.userRepo.Update(ctx, user); err != nil {
			return fmt.Errorf("error actualizando usuario base: %w", err)
		}
	}

	var tableName string
	var allowedFields map[string]bool
	switch roleID {
	case "role_participant":
		tableName = "participants"
		allowedFields = map[string]bool{
			"birthdate":               true,
			"dni":                     true,
			"address":                 true,
			"gender":                  true,
			"image_url":               true,
			"emergency_contact_name":  true,
			"emergency_contact_phone": true,
		}
	case "role_company":
		tableName = "companies"
		allowedFields = map[string]bool{
			"business_name": true,
			"cuit":          true,
		}
	case "role_teacher":
		tableName = "teachers"
		allowedFields = map[string]bool{
			"specialization": true,
			"is_external":    true,
		}
	case "role_volunteer":
		tableName = "volunteers"
		allowedFields = map[string]bool{}
	default:
		if len(data) > 0 {
			return errors.New("rol no soporta perfil extra o es inválido")
		}
		return nil
	}

	roleUpdateData := make(map[string]interface{})
	for k, v := range data {
		if allowedFields[k] {
			roleUpdateData[k] = v
		}
	}

	if len(roleUpdateData) == 0 {
		return nil // nothing to update
	}

	err = s.querier.UpdateRoleData(ctx, tableName, userID, roleUpdateData)
	if err != nil {
		return fmt.Errorf("error actualizando perfil: %w", err)
	}
	return nil
}
