package models

import (
	"context"
	"time"
)

var (
	TenantActiveState   int8 = 1
	TenantInactiveState int8 = 0
)

type Tenant struct {
	Id           uint64    `json:"id"`
	Uuid         string    `json:"uuid"`
	Name         string    `json:"name"`
	State        int8      `json:"state"`
	Timezone     string    `json:"timestamp"`
	CreatedTime  time.Time `json:"created_time"`
	ModifiedTime time.Time `json:"modified_time"`
}

type TenantRepository interface {
	Insert(ctx context.Context, u *Tenant) error
	UpdateById(ctx context.Context, u *Tenant) error
	GetById(ctx context.Context, id uint64) (*Tenant, error)
	GetByUuid(ctx context.Context, uuid string) (*Tenant, error)
	CheckIfExistsById(ctx context.Context, id uint64) (bool, error)
	CheckIfExistsByName(ctx context.Context, name string) (bool, error)
	InsertOrUpdateById(ctx context.Context, u *Tenant) error
	ListAllUuids(ctx context.Context) ([]string, error)
	ListAllIds(ctx context.Context) ([]uint64, error)
}
