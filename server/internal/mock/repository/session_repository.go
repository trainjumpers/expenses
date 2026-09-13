package mock_repository

import (
	"context"
	"errors"
	"expenses/internal/models"
	"sync"
	"time"
)

type MockSessionRepository struct {
	sessions map[string]models.Session
	nextId   int64
	mu       sync.RWMutex
	failures map[string]error
}

func NewMockSessionRepository() *MockSessionRepository {
	return &MockSessionRepository{
		sessions: make(map[string]models.Session),
		nextId:   1,
	}
}

// FailOn makes Create, RevokeAllForUser, ExpireByHash or DeleteExpired return err,
// so service tests can exercise failure paths.
func (m *MockSessionRepository) FailOn(method string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failures == nil {
		m.failures = make(map[string]error)
	}
	m.failures[method] = err
}

func (m *MockSessionRepository) Create(_ context.Context, userId int64, tokenHash string, expiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.failures["Create"]; err != nil {
		return err
	}
	m.sessions[tokenHash] = models.Session{
		Id:        m.nextId,
		UserId:    userId,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}
	m.nextId++
	return nil
}

func (m *MockSessionRepository) GetActiveByHash(_ context.Context, tokenHash string) (models.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, ok := m.sessions[tokenHash]
	if !ok || session.RevokedAt != nil || session.ExpiresAt.Before(time.Now()) {
		return models.Session{}, errors.New("session not found")
	}
	return session, nil
}

func (m *MockSessionRepository) Rotate(ctx context.Context, userId int64, oldHash, newHash string, expiresAt time.Time) error {
	m.RevokeByHash(ctx, oldHash)
	return m.Create(ctx, userId, newHash, expiresAt)
}

func (m *MockSessionRepository) RevokeByHash(_ context.Context, tokenHash string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok := m.sessions[tokenHash]
	if !ok || session.RevokedAt != nil {
		return 0, nil
	}
	now := time.Now()
	session.RevokedAt = &now
	m.sessions[tokenHash] = session
	return 1, nil
}

func (m *MockSessionRepository) RevokeAllForUser(_ context.Context, userId int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.failures["RevokeAllForUser"]; err != nil {
		return err
	}
	now := time.Now()
	for hash, session := range m.sessions {
		if session.UserId == userId && session.RevokedAt == nil {
			session.RevokedAt = &now
			m.sessions[hash] = session
		}
	}
	return nil
}

func (m *MockSessionRepository) ExpireByHash(_ context.Context, tokenHash string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.failures["ExpireByHash"]; err != nil {
		return 0, err
	}
	session, ok := m.sessions[tokenHash]
	if !ok {
		return 0, nil
	}
	session.ExpiresAt = time.Now().Add(-time.Minute)
	m.sessions[tokenHash] = session
	return 1, nil
}

func (m *MockSessionRepository) DeleteExpired(_ context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.failures["DeleteExpired"]; err != nil {
		return 0, err
	}
	var deleted int64
	for hash, session := range m.sessions {
		if session.ExpiresAt.Before(time.Now()) {
			delete(m.sessions, hash)
			deleted++
		}
	}
	return deleted, nil
}
