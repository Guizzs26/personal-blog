package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Guizzs26/personal-blog/internal/modules/identity/model"
	"github.com/google/uuid"
	"github.com/mdobak/go-xerrors"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (ur *PostgresUserRepository) Create(ctx context.Context, user model.User) (*model.User, error) {
	query := `
		INSERT INTO users
			(name, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, active, created_at, updated_at
	`

	var createdUser model.User
	err := ur.db.QueryRowContext(ctx, query,
		user.Name,
		user.Email,
		user.Password,
	).Scan(
		&createdUser.ID,
		&createdUser.Name,
		&createdUser.Email,
		&createdUser.Active,
		&createdUser.CreatedAt,
		&createdUser.UpdatedAt,
	)
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("repository: insert user: %v", err), 0)
	}

	return &createdUser, nil
}

func (ur *PostgresUserRepository) CreateFromGitHub(ctx context.Context, user model.User) (*model.User, error) {
	query := `
		INSERT INTO users
			(name, email, password, github_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, active, created_at, updated_at, github_id
	`

	var createdUser model.User
	err := ur.db.QueryRowContext(ctx, query,
		user.Name,
		user.Email,
		user.Password,
		user.GitHubID,
	).Scan(
		&createdUser.ID,
		&createdUser.Name,
		&createdUser.Email,
		&createdUser.Active,
		&createdUser.CreatedAt,
		&createdUser.UpdatedAt,
		&createdUser.GitHubID,
	)
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("repository: insert user (github): %v", err), 0)
	}

	return &createdUser, nil
}

// Generic method
func (ur *PostgresUserRepository) Update(ctx context.Context, user *model.User) error {
	const query = `
		UPDATE users
		SET name = $1, email = $2, password = $3, active = $4, github_id = $5, updated_at = NOW()
		WHERE id = $6
	`

	_, err := ur.db.ExecContext(ctx, query,
		user.Name,
		user.Email,
		user.Password,
		user.Active,
		user.GitHubID,
		user.ID,
	)

	if err != nil {
		return xerrors.WithStackTrace(fmt.Errorf("repository: update user: %v", err), 0)
	}

	return nil
}

func (ur *PostgresUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)
	`

	if err := ur.db.QueryRowContext(ctx, query, email).Scan(&exists); err != nil {
		return false, xerrors.WithStackTrace(fmt.Errorf("repository: check email existence: %v", err), 0)
	}

	return exists, nil
}

func (ur *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, name, email, password, active, created_at, updated_at
		FROM users
		WHERE email = $1 AND active = true
	`

	var user model.User
	err := ur.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("repository: find by email: %v", err), 0)
	}

	return &user, nil
}

func (ur *PostgresUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, name, email, password, active, created_at, updated_at
		FROM users
		WHERE id = $1 AND active = true
	`

	var user model.User
	err := ur.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("repository: find by id: %v", err), 0)
	}

	return &user, nil
}

func (ur *PostgresUserRepository) FindByGitHubID(ctx context.Context, gitHubID int64) (*model.User, error) {
	const query = `
		SELECT id, name, email, password, active, github_id, created_at, updated_at
		FROM users
		WHERE github_id = $1
	`

	var user model.User
	err := ur.db.QueryRowContext(ctx, query, gitHubID).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Active,
		&user.GitHubID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("repository: find by github_id: %v", err), 0)
	}

	return &user, nil
}
