package services

import (
	"context"

	"auth/internal/middlewares"
)

// canApprove dice si quien hace el pedido puede fijar el approval_status
// de un perfil del rol profileRole. Aprueban admin y dev (HU-12, HU-38);
// el voluntario, además, perfiles de participantes (HU-07). Sin esto, una
// cuenta pendiente podía aprobarse sola con PUT sobre su propio perfil.
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
