package main

import (
	"sync"
)

// VideoStorage управляет хранилищем видео и сессий
type VideoStorage struct {
	Sessions map[string]*Session
	Videos   map[string][]byte
	mu       sync.RWMutex
}

// NewVideoStorage создает новое хранилище
func NewVideoStorage() *VideoStorage {
	return &VideoStorage{
		Sessions: make(map[string]*Session),
		Videos:   make(map[string][]byte),
	}
}

// CreateOrGetSession получает сессию или создает новую
func (vs *VideoStorage) CreateOrGetSession(sessionID string) *Session {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	if session, exists := vs.Sessions[sessionID]; exists {
		return session
	}

	session := NewSession(sessionID)
	vs.Sessions[sessionID] = session
	return session
}

// GetSession получает сессию по ID
func (vs *VideoStorage) GetSession(sessionID string) *Session {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	return vs.Sessions[sessionID]
}

// RemoveSession удаляет сессию
func (vs *VideoStorage) RemoveSession(sessionID string) {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	delete(vs.Sessions, sessionID)
}

// StoreVideo сохраняет видео
func (vs *VideoStorage) StoreVideo(sessionID string, data []byte) {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	vs.Videos[sessionID] = data
}

// GetVideo получает видео по ID сессии
func (vs *VideoStorage) GetVideo(sessionID string) []byte {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	return vs.Videos[sessionID]
}

// DeleteVideo удаляет видео
func (vs *VideoStorage) DeleteVideo(sessionID string) {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	delete(vs.Videos, sessionID)
}

// SessionCount возвращает количество активных сессий
func (vs *VideoStorage) SessionCount() int {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	return len(vs.Sessions)
}
