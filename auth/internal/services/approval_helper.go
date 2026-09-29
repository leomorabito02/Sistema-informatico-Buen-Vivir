package services

import (
	"context"

	"auth/internal/middlewares"
)

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
