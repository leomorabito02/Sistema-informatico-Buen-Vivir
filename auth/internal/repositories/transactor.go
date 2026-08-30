package repositories

import "context"

// Transactor abstracts GORM transaction management.
type Transactor interface {
	// RunInTx executes fn within a DB transaction. Rolls back on error.
	RunInTx(ctx context.Context, fn func(tx Tx) error) error
}

// Tx is a minimal write-only DB session used inside a transaction.
type Tx interface {
	Create(value any) error
}

// RoleProfileQuerier handles dynamic role-table queries
// (teachers, participants, companies, volunteers).
type RoleProfileQuerier interface {
	GetApprovalStatus(ctx context.Context, table, userID string) (string, error)
	GetIsExternal(ctx context.Context, userID string) (bool, error)
	GetRoleData(ctx context.Context, table, userID string) (map[string]any, error)
	UpdateRoleData(ctx context.Context, table, userID string, data map[string]any) error
}
