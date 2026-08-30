package repositories

import (
	"context"

	"gorm.io/gorm"
)

type gormTransactor struct {
	db *gorm.DB
}

func NewGormTransactor(db *gorm.DB) Transactor {
	return &gormTransactor{db: db}
}

func (g *gormTransactor) RunInTx(ctx context.Context, fn func(tx Tx) error) error {
	return g.db.WithContext(ctx).Transaction(func(dbTx *gorm.DB) error {
		return fn(&gormTx{tx: dbTx})
	})
}

type gormTx struct {
	tx *gorm.DB
}

func (g *gormTx) Create(value any) error {
	return g.tx.Create(value).Error
}

type gormRoleProfileQuerier struct {
	db *gorm.DB
}

func NewGormRoleProfileQuerier(db *gorm.DB) RoleProfileQuerier {
	return &gormRoleProfileQuerier{db: db}
}

func (q *gormRoleProfileQuerier) GetApprovalStatus(ctx context.Context, table, userID string) (string, error) {
	var status string
	err := q.db.WithContext(ctx).Table(table).Where("user_id = ?", userID).Select("approval_status").Scan(&status).Error
	return status, err
}

func (q *gormRoleProfileQuerier) GetIsExternal(ctx context.Context, userID string) (bool, error) {
	var isExt bool
	err := q.db.WithContext(ctx).Table("teachers").Where("user_id = ?", userID).Select("is_external").Scan(&isExt).Error
	return isExt, err
}

func (q *gormRoleProfileQuerier) GetRoleData(ctx context.Context, table, userID string) (map[string]any, error) {
	var roleData map[string]any
	err := q.db.WithContext(ctx).Table(table).Where("user_id = ?", userID).Take(&roleData).Error
	return roleData, err
}

func (q *gormRoleProfileQuerier) UpdateRoleData(ctx context.Context, table, userID string, data map[string]any) error {
	return q.db.WithContext(ctx).Table(table).Where("user_id = ?", userID).Updates(data).Error
}
