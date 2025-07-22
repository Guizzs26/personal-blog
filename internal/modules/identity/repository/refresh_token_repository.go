package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Guizzs26/personal-blog/internal/modules/identity/model"
	"github.com/google/uuid"
	"github.com/mdobak/go-xerrors"
)

type PostgresRefreshTokenRepository struct {
	db *sql.DB
}

func NewPostgresRefreshTokenRepository(db *sql.DB) *PostgresRefreshTokenRepository {
	return &PostgresRefreshTokenRepository{db: db}
}

func (prr *PostgresRefreshTokenRepository) Save(ctx context.Context, refresh *model.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, user_agent, ip_address, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := prr.db.QueryRowContext(ctx, query,
		refresh.UserID,
		refresh.TokenHash,
		refresh.UserAgent,
		refresh.IPAddress,
		refresh.CreatedAt,
		refresh.ExpiresAt,
	).Scan(&refresh.ID)

	if err != nil {
		return xerrors.WithStackTrace(fmt.Errorf("repository: insert refresh token: %v", err), 0)
	}

	return nil
}

func (prr *PostgresRefreshTokenRepository) RevokeByID(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked_at = $1 WHERE id = $2`

	_, err := prr.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return xerrors.WithStackTrace(fmt.Errorf("repository: revoke refresh token by id: %v", err), 0)
	}

	return nil
}

func (prr *PostgresRefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < $1
		OR revoked_at IS NOT NULL
	`

	_, err := prr.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return xerrors.WithStackTrace(fmt.Errorf("repository: delete expired refresh tokens: %v", err), 0)
	}

	return nil
}

func (prr *PostgresRefreshTokenRepository) FindByHash(ctx context.Context, hash string) (*model.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, user_agent, ip_address, created_at, expires_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = $1
		LIMIT 1
	`

	var token model.RefreshToken
	err := prr.db.QueryRowContext(ctx, query, hash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.UserAgent,
		&token.IPAddress,
		&token.CreatedAt,
		&token.ExpiresAt,
		&token.RevokedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}

	if err != nil {
		return nil, xerrors.WithStackTrace(fmt.Errorf("repository: find refresh token by hash: %v", err), 0)
	}

	return &token, nil
}
