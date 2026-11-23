package services

import (
	"context"
	"errors"

	"auth/internal/middlewares"
)

// errCreateForOtherUser: un usuario sin rol de gestión quiso crear un
// perfil a nombre de otra cuenta.
var errCreateForOtherUser = errors.New("forbidden: access denied to create a profile for another user")

// checkCreateProfileOwnership: admin, dev y los staffRoles crean perfiles para cualquier
// cuenta (el voluntario crea participantes, HU-07); el resto, solo el de su propia cuenta.
func checkCreateProfileOwnership(ctx context.Context, targetUserID string, staffRoles ...string) error {
	if isAdminRequest(ctx) {
		return nil
	}
	reqRole, _ := ctx.Value(middlewares.RoleIDKey).(string)
	for _, r := range staffRoles {
		if reqRole == r {
			return nil
		}
	}
	reqUserID, _ := ctx.Value(middlewares.UserIDKey).(string)
	if reqUserID == "" || reqUserID != targetUserID {
		return errCreateForOtherUser
	}
	return nil
}

// canApprove dice si quien pide puede fijar el approval_status de un perfil de
// profileRole: admin y dev (HU-12, HU-38) y el voluntario para participantes (HU-07).
func canApprove(ctx context.Context, profileRole string) bool {
	reqRole, _ := ctx.Value(middlewares.RoleIDKey).(string)
	switch reqRole {
	case "role_admin", "role_dev":
		return true
	case "role_volunteer":
		return profileRole == "role_participant"
	}
	return false
}

// isAdminRequest dice si el pedido lo hace un admin o un dev.
func isAdminRequest(ctx context.Context) bool {
	reqRole, _ := ctx.Value(middlewares.RoleIDKey).(string)
	return reqRole == "role_admin" || reqRole == "role_dev"
}

// resolveApprover is a helper function to reduce cognitive complexity across
// different services when determining the user ID that approved a request.
func resolveApprover(ctx context.Context, reqApprovedBy string) *string {
	if reqApprovedBy != "" {
		return &reqApprovedBy
	}
	if reqUserID, ok := ctx.Value(middlewares.UserIDKey).(string); ok && reqUserID != "" {
		return &reqUserID
	}
	return nil
}
