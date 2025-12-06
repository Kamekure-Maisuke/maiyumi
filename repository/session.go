package repository

import "sync"

type SessionRepository interface {
	Set(sessionID, username string)
	Get(sessionID string) (string, bool)
	Delete(sessionID string)
}

type sessionRepository struct {
	sessions map[string]string
	mu       sync.RWMutex
}

func NewSessionRepository() SessionRepository {
	return &sessionRepository{
		sessions: make(map[string]string),
	}
}

func (r *sessionRepository) Set(sessionID, username string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[sessionID] = username
}

func (r *sessionRepository) Get(sessionID string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	username, ok := r.sessions[sessionID]
	return username, ok
}

func (r *sessionRepository) Delete(sessionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, sessionID)
}
