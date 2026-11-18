package audience_test

import (
	"context"
	"errors"
	"testing"

	"buenvivir/notifications/internal/audience"
	"buenvivir/notifications/internal/clients"
	"buenvivir/notifications/internal/rules"
)

type mockAuthClient struct {
	participantUserID string
	getParticipantErr error
	roleUserIDs       []string
	getRoleUserIDsErr error
}

func (m *mockAuthClient) GetUserDeviceTokens(ctx context.Context, userID string) ([]string, error) {
	return nil, nil
}
func (m *mockAuthClient) RevokeDeviceTokens(ctx context.Context, userID string, tokens []string) error {
	return nil
}
func (m *mockAuthClient) GetCampaignRecipients(ctx context.Context) ([]clients.RecipientContact, error) {
	return nil, nil
}
func (m *mockAuthClient) IsUserEmailVerified(ctx context.Context, userID string) (bool, error) {
	return true, nil
}
func (m *mockAuthClient) GetUserContact(ctx context.Context, userID string) (*clients.RecipientContact, error) {
	return nil, nil
}
func (m *mockAuthClient) GetActiveUserIDsByRoles(ctx context.Context, roleIDs []string) ([]string, error) {
	if m.getRoleUserIDsErr != nil {
		return nil, m.getRoleUserIDsErr
	}
	return m.roleUserIDs, nil
}
func (m *mockAuthClient) GetParticipantUserID(ctx context.Context, participantID string) (string, error) {
	if m.getParticipantErr != nil {
		return "", m.getParticipantErr
	}
	return m.participantUserID, nil
}
func (m *mockAuthClient) IssueActionLink(ctx context.Context, requestID, userID, actionType string) (string, error) {
	return "https://example.com/action", nil
}
func (m *mockAuthClient) Close() error {
	return nil
}

func TestAudienceResolver(t *testing.T) {
	ctx := context.Background()

	t.Run("AudienceIndividual", func(t *testing.T) {
		r := audience.NewResolver(nil)
		plan := rules.NotificationPlan{
			AudienceType:     rules.AudienceIndividual,
			RecipientUserIDs: []string{"usr-1", "usr-2"},
		}
		uids, err := r.ResolveUserIDs(ctx, plan)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(uids) != 2 || uids[0] != "usr-1" {
			t.Errorf("unexpected uids: %v", uids)
		}
	})

	t.Run("AudienceParticipant_Success", func(t *testing.T) {
		mock := &mockAuthClient{participantUserID: "usr-from-part"}
		r := audience.NewResolver(mock)
		plan := rules.NotificationPlan{
			AudienceType:  rules.AudienceParticipant,
			ParticipantID: "part-10",
		}
		uids, err := r.ResolveUserIDs(ctx, plan)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(uids) != 1 || uids[0] != "usr-from-part" {
			t.Errorf("expected usr-from-part, got %v", uids)
		}
	})

	t.Run("AudienceParticipant_Empty", func(t *testing.T) {
		r := audience.NewResolver(nil)
		plan := rules.NotificationPlan{AudienceType: rules.AudienceParticipant}
		_, err := r.ResolveUserIDs(ctx, plan)
		if err == nil {
			t.Fatal("expected error on empty participantID")
		}
	})

	t.Run("AudienceRole_Success", func(t *testing.T) {
		mock := &mockAuthClient{roleUserIDs: []string{"admin-1", "admin-2"}}
		r := audience.NewResolver(mock)
		adminRole := "role_admin"
		plan := rules.NotificationPlan{
			AudienceType: rules.AudienceRole,
			TargetRole:   &adminRole,
		}
		uids, err := r.ResolveUserIDs(ctx, plan)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(uids) != 2 || uids[0] != "admin-1" {
			t.Errorf("unexpected uids: %v", uids)
		}
	})

	t.Run("AudienceRole_Error", func(t *testing.T) {
		mock := &mockAuthClient{getRoleUserIDsErr: errors.New("auth unavailable")}
		r := audience.NewResolver(mock)
		adminRole := "role_admin"
		plan := rules.NotificationPlan{
			AudienceType: rules.AudienceRole,
			TargetRole:   &adminRole,
		}
		_, err := r.ResolveUserIDs(ctx, plan)
		if err == nil {
			t.Fatal("expected error on auth failure")
		}
	})
}

