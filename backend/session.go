package main

import (
	"sync"
	"time"
)

// Session представляет сессию совместного просмотра
type Session struct {
	ID           string
	Peers        map[string]*Peer
	VideoData    []byte
	Playing      bool
	CurrentTime  float64
	LastUpdate   time.Time
	mu           sync.RWMutex
	CreatedAt    time.Time
}

// NewSession создает новую сессию
func NewSession(id string) *Session {
	return &Session{
		ID:        id,
		Peers:     make(map[string]*Peer),
		VideoData: nil,
		Playing:   false,
		CurrentTime: 0,
		CreatedAt: time.Now(),
	}
}

// AddPeer добавляет клиента в сессию
func (s *Session) AddPeer(peer *Peer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Peers[peer.ID] = peer
}

// RemovePeer удаляет клиента из сессии
func (s *Session) RemovePeer(peerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Peers, peerID)
}

// GetPeerIDs возвращает список ID подключенных клиентов
func (s *Session) GetPeerIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.Peers))
	for id := range s.Peers {
		ids = append(ids, id)
	}
	return ids
}

// BroadcastSync отправляет синхронизацию всем клиентам
func (s *Session) BroadcastSync(excludePeerID string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	msg := map[string]interface{}{
		"type":        "sync",
		"playing":     s.Playing,
		"currentTime": s.CurrentTime,
		"timestamp":   time.Now().Unix(),
	}

	for id, peer := range s.Peers {
		if id != excludePeerID {
			peer.Send(msg)
		}
	}
}

// BroadcastPeerUpdate уведомляет о изменении списка клиентов
func (s *Session) BroadcastPeerUpdate() {
	s.mu.RLock()
	peerIDs := make([]string, 0, len(s.Peers))
	for id := range s.Peers {
		peerIDs = append(peerIDs, id)
	}
	peers := s.Peers
	s.mu.RUnlock()

	msg := map[string]interface{}{
		"type":  "peerUpdate",
		"peers": peerIDs,
	}

	for _, peer := range peers {
		peer.Send(msg)
	}
}

// IsEmpty проверяет, нет ли клиентов в сессии
func (s *Session) IsEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Peers) == 0
}

// GetPeerCount возвращает количество клиентов
func (s *Session) GetPeerCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Peers)
}
