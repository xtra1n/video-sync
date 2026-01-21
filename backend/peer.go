package main

import (
	"log"

	"github.com/gorilla/websocket"
)

// Peer представляет подключенного клиента
type Peer struct {
	ID        string
	SessionID string
	Conn      *websocket.Conn
	SendChan  chan map[string]interface{}
	mu        chan struct{} // для синхронизации
	closed    bool
}

// NewPeer создает нового клиента
func NewPeer(id, sessionID string, conn *websocket.Conn) *Peer {
	return &Peer{
		ID:        id,
		SessionID: sessionID,
		Conn:      conn,
		SendChan:  make(chan map[string]interface{}, 10),
		mu:        make(chan struct{}, 1),
		closed:    false,
	}
}

// Send отправляет сообщение клиенту
func (p *Peer) Send(msg map[string]interface{}) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Send error for peer %s: channel closed", p.ID)
		}
	}()
	
	if p.closed {
		return
	}
	
	select {
	case p.SendChan <- msg:
	default:
		log.Printf("Send channel full for peer %s", p.ID)
	}
}

// HandleMessages читает и обрабатывает сообщения от клиента
func (p *Peer) HandleMessages(handler func(map[string]interface{})) {
	defer close(p.SendChan)

	for {
		var msg map[string]interface{}
		err := p.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for peer %s: %v", p.ID, err)
			}
			return
		}
		handler(msg)
	}
}

// SendLoop отправляет сообщения из очереди клиенту
func (p *Peer) SendLoop() {
	defer func() {
		p.closed = true
		p.Conn.Close()
	}()

	for msg := range p.SendChan {
		err := p.Conn.WriteJSON(msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Failed to send to peer %s: %v", p.ID, err)
			}
			return
		}
	}
}

// Close закрывает соединение с клиентом
func (p *Peer) Close() error {
	p.closed = true
	return p.Conn.Close()
}
