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

type UserRepo struct {
	dbpool *pgxpool.Pool
}

func NewUserRepo(dbpool *pgxpool.Pool) *UserRepo {
	return &UserRepo{
		dbpool: dbpool,
	}
}

func (r *UserRepo) Insert(ctx context.Context, m *models.User) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
    INSERT INTO users (
        uuid, tenant_id, email, first_name, last_name, password_algorithm, password_hash, state,
		role_id, timezone, created_time, modified_time, salt, was_email_activated,
		pr_access_code, pr_expiry_time
    ) VALUES (
        $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
    )`

	_, err := r.dbpool.Exec(ctx, query, m.Uuid, m.TenantId, m.Email, m.FirstName, m.LastName, m.PasswordAlgorithm, m.PasswordHash, m.State, m.RoleId, m.Timezone, m.CreatedTime, m.ModifiedTime, m.Salt, m.WasEmailActivated, m.PrAccessCode, m.PrExpiryTime)
	if err != nil {
		log.Println("UserRepo|Insert|err", err)
		return err
	}
	return nil
}

func (r *UserRepo) UpdateById(ctx context.Context, m *models.User) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
    UPDATE
        users
    SET
        tenant_id = $1, email = $2, first_name = $3, last_name = $4, password_algorithm = $5, password_hash = $6, state = $7,
		role_id = $8, timezone = $9, created_time = $10, modified_time = $11, salt = $12, was_email_activated = $13,
		pr_access_code = $14, pr_expiry_time = $15
    WHERE
        id = $16`

	err := r.dbpool.QueryRow(ctx, query).Scan(&m)
	if err != nil {
		log.Println("UserRepo|UpdateById|err", err)
		return err
	}
	return nil
}

//
func (r *UserRepo) UpdateByEmail(ctx context.Context, m *models.User) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
    UPDATE
        users
    SET
        tenant_id = $1, email = $2, first_name = $3, last_name = $4, password_algorithm = $5, password_hash = $6, state = $7,
		role_id = $8, timezone = $9, created_time = $10, modified_time = $11, salt = $12, was_email_activated = $13,
		pr_access_code = $14, pr_expiry_time = $15
    WHERE
        email = $2`

	err := r.dbpool.QueryRow(ctx, query).Scan(&m)
	if err != nil {
		log.Println("UserRepo|UpdateByEmail|err", err)
		return err
	}
	return nil
}

func (r *UserRepo) GetById(ctx context.Context, id uint64) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	m := new(models.User)

	query := `
    SELECT
        id, uuid, tenant_id, email, first_name, last_name, password_algorithm, password_hash, state,
		role_id, timezone, created_time, modified_time, salt, was_email_activated, pr_access_code, pr_expiry_time
    FROM
        users
    WHERE
        id = $1`

	err := r.dbpool.QueryRow(ctx, query, id).Scan(&m)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		} else {
			log.Println("UserRepo|GetById|err", err)
			return nil, err
		}
	}
	return m, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	m := new(models.User)

	query := `
    SELECT
        id, uuid, tenant_id, email, first_name, last_name, password_algorithm,
		password_hash, state, role_id, timezone, created_time, modified_time,
		salt, was_email_activated, pr_access_code, pr_expiry_time
    FROM
        users
    WHERE
        email = $1`

	err := r.dbpool.QueryRow(ctx, query, email).Scan(
		&m.Id, &m.Uuid, &m.TenantId, &m.Email, &m.FirstName, &m.LastName,
		&m.PasswordAlgorithm, &m.PasswordHash, &m.State, &m.RoleId, &m.Timezone,
		&m.CreatedTime, &m.ModifiedTime, &m.Salt, &m.WasEmailActivated,
		&m.PrAccessCode, &m.PrExpiryTime)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return m, nil
}

func (r *UserRepo) CheckIfExistsById(ctx context.Context, id uint64) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var exists int

	query := `
    SELECT
        1
    FROM
        users
    WHERE
        id = $1`

	err := r.dbpool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		} else {
			log.Println("UserRepo|CheckIfExistsById|err", err)
			return false, err
		}
	}
	return exists == 1, nil
}

func (r *UserRepo) CheckIfExistsByEmail(ctx context.Context, email string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var exists int

	query := `
    SELECT
        1
    FROM
        users
    WHERE
        email = $1`

	err := r.dbpool.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		} else {
			log.Println("UserRepo|CheckIfExistsByEmail|err", err)
			return false, err
		}
	}
	return exists == 1, nil
}

func (r *UserRepo) InsertOrUpdateById(ctx context.Context, m *models.User) error {
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

func (r *UserRepo) InsertOrUpdateByEmail(ctx context.Context, m *models.User) error {
	if m.Id == 0 {
		return r.Insert(ctx, m)
	}

	doesExist, err := r.CheckIfExistsByEmail(ctx, m.Email)
	if err != nil {
		return err
	}

	if doesExist == false {
		return r.Insert(ctx, m)
	}
	return r.UpdateByEmail(ctx, m)
}
