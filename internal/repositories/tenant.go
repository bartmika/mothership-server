package repositories

import (
	"context"
	// "database/sql"
	"log"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/bartmika/mothership-server/internal/models"
)

type TenantRepo struct {
	dbpool *pgxpool.Pool
}

func NewTenantRepo(dbpool *pgxpool.Pool) *TenantRepo {
	return &TenantRepo{
		dbpool: dbpool,
	}
}

func (r *TenantRepo) Insert(ctx context.Context, m *models.Tenant) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
    INSERT INTO tenants (
        uuid, name, state, timezone, created_time, modified_time

    ) VALUES (
        $1, $2, $3, $4, $5, $6
    )
    `

	_, err := r.dbpool.Exec(ctx, query, m.Uuid, m.Name, m.State, m.Timezone, m.CreatedTime, m.ModifiedTime)
	if err != nil {
		log.Println("TenantRepo|Insert|err", err)
		return err
	}
	return nil
}

func (r *TenantRepo) UpdateById(ctx context.Context, m *models.Tenant) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
    UPDATE
        tenants
    SET
        name = $1, state = $2, timezone = $3, created_time = $4, modified_time = $5
    WHERE
        id = $6
    `

	err := r.dbpool.QueryRow(ctx, query).Scan(&m)
	if err != nil {
		log.Println("TenantRepo|UpdateById|err", err)
		return err
	}
	return nil
}

func (r *TenantRepo) GetById(ctx context.Context, id uint64) (*models.Tenant, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	m := new(models.Tenant)

	query := `
    SELECT
        id, uuid, name, state, timezone, created_time, modified_time
    FROM
        tenants
    WHERE
        id = $1
    `
	err := r.dbpool.QueryRow(ctx, query, id).Scan(&m.Id, &m.Uuid, &m.Name, &m.State, &m.Timezone, &m.CreatedTime, &m.ModifiedTime)
	if err != nil {
		log.Println("TenantRepo|GetById|err", err)
		return nil, err
	}
	return m, nil
}

func (r *TenantRepo) GetByUuid(ctx context.Context, uid string) (*models.Tenant, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	m := new(models.Tenant)

	query := `
    SELECT
        id, uuid, name, state, timezone, created_time, modified_time
    FROM
        tenants
    WHERE
        uuid = $1
    `
	err := r.dbpool.QueryRow(ctx, query, uid).Scan(&m.Id, &m.Uuid, &m.Name, &m.State, &m.Timezone, &m.CreatedTime, &m.ModifiedTime)
	if err != nil {
		log.Println("TenantRepo|GetByUuid|err", err)
		return nil, err
	}
	return m, nil
}

func (r *TenantRepo) CheckIfExistsById(ctx context.Context, id uint64) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var exists int

	query := `
    SELECT
        1
    FROM
        tenants
    WHERE
        id = $1
    `

	err := r.dbpool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		} else {
			log.Println("TenantRepo|CheckIfExistsById|err", err)
			return false, err
		}
	}
	return exists == 1, nil
}

func (r *TenantRepo) CheckIfExistsByName(ctx context.Context, name string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var exists int

	query := `
    SELECT
        1
    FROM
        tenants
    WHERE
        name = $1
    `

	err := r.dbpool.QueryRow(ctx, query, name).Scan(&exists)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		} else {
			log.Println("TenantRepo|CheckIfExistsByName|err", err)
			return false, err
		}
	}
	return exists == 1, nil
}

func (r *TenantRepo) InsertOrUpdateById(ctx context.Context, m *models.Tenant) error {
	if m.Id == 0 {
		return r.Insert(ctx, m)
	}

	doesExist, err := r.CheckIfExistsById(ctx, m.Id)
	if err != nil {
		return err
	}

	if doesExist == false {
		return r.Insert(ctx, m)
	}
	return r.UpdateById(ctx, m)
}

func (r *TenantRepo) ListAllUuids(ctx context.Context) ([]string, error) {
	var uuids []string

	query := `SELECT uuid FROM tenants ORDER BY (id) ASC`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.dbpool.Query(ctx, query)
	if err != nil {
		return uuids, err
	}
	defer rows.Close()

	for rows.Next() {
		var r string
		err = rows.Scan(&r)
		if err != nil {
			return uuids, err
		}
		uuids = append(uuids, r)
	}

	// Any errors encountered by rows.Next or rows.Scan will be returned here
	if rows.Err() != nil {
		return uuids, rows.Err()
	}

	return uuids, nil
}

func (r *TenantRepo) ListAllIds(ctx context.Context) ([]uint64, error) {
	var ids []uint64

	query := `SELECT id FROM tenants ORDER BY (id) ASC`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.dbpool.Query(ctx, query)
	if err != nil {
		return ids, err
	}
	defer rows.Close()

	for rows.Next() {
		var r uint64
		err = rows.Scan(&r)
		if err != nil {
			return ids, err
		}
		ids = append(ids, r)
	}

	// Any errors encountered by rows.Next or rows.Scan will be returned here
	if rows.Err() != nil {
		return ids, rows.Err()
	}

	return ids, nil
}
