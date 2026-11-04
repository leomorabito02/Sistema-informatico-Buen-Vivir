package audience

import (
	"context"
	"fmt"
	"strings"

	"buenvivir/notifications/internal/clients"
	"buenvivir/notifications/internal/rules"
)

// Resolver resolves recipient user IDs from a NotificationPlan.
type Resolver interface {
	ResolveUserIDs(ctx context.Context, plan rules.NotificationPlan) ([]string, error)
}

type authAudienceResolver struct {
	authClient clients.AuthClient
}

// NewResolver constructs an AudienceResolver backed by the Auth gRPC client.
func NewResolver(authClient clients.AuthClient) Resolver {
	return &authAudienceResolver{authClient: authClient}
}

func (r *authAudienceResolver) ResolveUserIDs(ctx context.Context, plan rules.NotificationPlan) ([]string, error) {
	switch plan.AudienceType {
	case rules.AudienceIndividual:
		return plan.RecipientUserIDs, nil

	case rules.AudienceParticipant:
		if strings.TrimSpace(plan.ParticipantID) == "" {
			return nil, fmt.Errorf("plan requires participant_id")
		}
		if r.authClient == nil {
			return []string{plan.ParticipantID}, nil
		}
		userID, err := r.authClient.GetParticipantUserID(ctx, plan.ParticipantID)
		if err != nil {
			return nil, fmt.Errorf("could not resolve participant %s: %w", plan.ParticipantID, err)
		}
		if strings.TrimSpace(userID) == "" {
			return nil, fmt.Errorf("resolved user_id for participant %s is empty", plan.ParticipantID)
		}
		return []string{userID}, nil

	case rules.AudienceRole:
		if plan.TargetRole == nil || strings.TrimSpace(*plan.TargetRole) == "" {
			return nil, fmt.Errorf("plan requires target_role")
		}
		if r.authClient == nil {
			return nil, fmt.Errorf("auth client unavailable to resolve role")
		}
		userIDs, err := r.authClient.GetActiveUserIDsByRoles(ctx, []string{*plan.TargetRole})
		if err != nil {
			return nil, fmt.Errorf("could not resolve role %s: %w", *plan.TargetRole, err)
		}
		return userIDs, nil

	default:
		return plan.RecipientUserIDs, nil
	}
}

