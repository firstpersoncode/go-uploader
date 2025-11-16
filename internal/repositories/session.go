package repositories

import (
	"fmt"
	"sync"

	"firstpersoncode/go-uploader/domain"
	"firstpersoncode/go-uploader/internal/util"
)

type sessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]*domain.Session
}

func NewSessionRepository() domain.SessionRepository {
	return &sessionRepository{
		sessions: make(map[string]*domain.Session),
	}
}

func (r *sessionRepository) Save(session *domain.Session) (*domain.Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if session.UserID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	session.ID = util.GenerateRandomID()
	session.RefreshToken = util.GenerateRandomID()

	r.sessions[session.ID] = session
	return session, nil
}

func (r *sessionRepository) FindByID(id string) (*domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	return session, nil
}
