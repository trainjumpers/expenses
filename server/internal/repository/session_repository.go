package repository

import (
	"context"
	"expenses/internal/config"
	"expenses/internal/models"
	database "expenses/pkg/database/manager"
	"fmt"
	"time"
)

type SessionRepositoryInterface interface {
	Create(ctx context.Context, userId int64, tokenHash string, expiresAt time.Time) error
	GetActiveByHash(ctx context.Context, tokenHash string) (models.Session, error)
	Rotate(ctx context.Context, userId int64, oldHash, newHash string, expiresAt time.Time) error
	RevokeByHash(ctx context.Context, tokenHash string) (int64, error)
	RevokeAllForUser(ctx context.Context, userId int64) error
	ExpireByHash(ctx context.Context, tokenHash string) (int64, error)
	DeleteExpired(ctx context.Context) (int64, error)
}

type SessionRepository struct {
	db        database.DatabaseManager
	schema    string
	tableName string
}

func NewSessionRepository(db database.DatabaseManager, cfg *config.Config) SessionRepositoryInterface {
	return &SessionRepository{
		db:        db,
		schema:    cfg.DBSchema,
		tableName: "session",
	}
}

func (r *SessionRepository) Create(ctx context.Context, userId int64, tokenHash string, expiresAt time.Time) error {
	query := fmt.Sprintf(`INSERT INTO %s.%s (user_id, token_hash, expires_at) VALUES ($1, $2, $3);`, r.schema, r.tableName)
	_, err := r.db.ExecuteQuery(ctx, query, userId, tokenHash, expiresAt)
	return err
}

func (r *SessionRepository) GetActiveByHash(ctx context.Context, tokenHash string) (models.Session, error) {
	var session models.Session
	query := fmt.Sprintf(`
		SELECT id, user_id, token_hash, expires_at, revoked_at
		FROM %s.%s
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW();`, r.schema, r.tableName)
	err := r.db.FetchOne(ctx, query, tokenHash).Scan(&session.Id, &session.UserId, &session.TokenHash, &session.ExpiresAt, &session.RevokedAt)
	if err != nil {
		return models.Session{}, err
	}
	return session, nil
}

func (r *SessionRepository) Rotate(ctx context.Context, userId int64, oldHash, newHash string, expiresAt time.Time) error {
	return r.db.WithTxn(ctx, func(txCtx context.Context) error {
		revokeQuery := fmt.Sprintf(`UPDATE %s.%s SET revoked_at = NOW() WHERE token_hash = $1 AND revoked_at IS NULL;`, r.schema, r.tableName)
		if _, err := r.db.ExecuteQuery(txCtx, revokeQuery, oldHash); err != nil {
			return err
		}
		insertQuery := fmt.Sprintf(`INSERT INTO %s.%s (user_id, token_hash, expires_at) VALUES ($1, $2, $3);`, r.schema, r.tableName)
		_, err := r.db.ExecuteQuery(txCtx, insertQuery, userId, newHash, expiresAt)
		return err
	})
}

func (r *SessionRepository) RevokeByHash(ctx context.Context, tokenHash string) (int64, error) {
	query := fmt.Sprintf(`UPDATE %s.%s SET revoked_at = NOW() WHERE token_hash = $1 AND revoked_at IS NULL;`, r.schema, r.tableName)
	return r.db.ExecuteQuery(ctx, query, tokenHash)
}

func (r *SessionRepository) RevokeAllForUser(ctx context.Context, userId int64) error {
	query := fmt.Sprintf(`UPDATE %s.%s SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL;`, r.schema, r.tableName)
	_, err := r.db.ExecuteQuery(ctx, query, userId)
	return err
}

func (r *SessionRepository) ExpireByHash(ctx context.Context, tokenHash string) (int64, error) {
	query := fmt.Sprintf(`UPDATE %s.%s SET expires_at = NOW() - INTERVAL '1 minute' WHERE token_hash = $1;`, r.schema, r.tableName)
	return r.db.ExecuteQuery(ctx, query, tokenHash)
}

func (r *SessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := fmt.Sprintf(`DELETE FROM %s.%s WHERE expires_at < NOW();`, r.schema, r.tableName)
	return r.db.ExecuteQuery(ctx, query)
}
