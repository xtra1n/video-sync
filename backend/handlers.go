package main

import (
	"encoding/base64"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	peerCounter int64
)

// handleWebSocket обрабатывает WebSocket соединение
func handleWebSocket(w http.ResponseWriter, r *http.Request, storage *VideoStorage) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	peerID := "peer_" + time.Now().Format("150405") + "_" + string(rune(atomic.AddInt64(&peerCounter, 1)))

	var joinMsg map[string]interface{}
	err = conn.ReadJSON(&joinMsg)
	if err != nil {
		log.Printf("Failed to read join message: %v", err)
		conn.Close()
		return
	}

	sessionID, ok := joinMsg["sessionId"].(string)
	if !ok || sessionID == "" {
		log.Printf("Invalid session ID in join message")
		conn.Close()
		return
	}

	session := storage.CreateOrGetSession(sessionID)
	peer := NewPeer(peerID, sessionID, conn)
	session.AddPeer(peer)

	log.Printf("Peer %s joined session %s (total peers: %d)", peerID, sessionID, session.GetPeerCount())

	peer.Send(map[string]interface{}{
		"type":   "joined",
		"peerID": peerID,
	})

	if video := storage.GetVideo(sessionID); video != nil {
		peer.Send(map[string]interface{}{
			"type": "videoData",
			"data": base64.StdEncoding.EncodeToString(video),
		})
	}

	session.mu.RLock()
	peer.Send(map[string]interface{}{
		"type":        "sync",
		"playing":     session.Playing,
		"currentTime": session.CurrentTime,
	})
	session.mu.RUnlock()

	session.BroadcastPeerUpdate()

	go peer.SendLoop()
	go handlePeerMessages(peer, session, storage)

	defer func() {
		session.RemovePeer(peerID)
		log.Printf("Peer %s left session %s (remaining: %d)", peerID, sessionID, session.GetPeerCount())

		session.BroadcastPeerUpdate()

		if session.IsEmpty() {
			storage.DeleteVideo(sessionID)
			storage.RemoveSession(sessionID)
			log.Printf("Session %s removed (empty)", sessionID)
		}
	}()

	<-time.After(time.Hour * 24)
}

// handlePeerMessages обрабатывает сообщения от клиента
func handlePeerMessages(peer *Peer, session *Session, storage *VideoStorage) {
	peer.HandleMessages(func(msg map[string]interface{}) {
		msgType, ok := msg["type"].(string)
		if !ok {
			log.Printf("Invalid message type from peer %s", peer.ID)
			return
		}

		switch msgType {
		case "sync":
			handleSyncMessage(peer, session, msg)

		case "videoData":
			handleVideoDataMessage(peer, session, storage, msg)

		default:
			log.Printf("Unknown message type: %s", msgType)
		}
	})
}

// handleSyncMessage обрабатывает синхронизацию воспроизведения
func handleSyncMessage(peer *Peer, session *Session, msg map[string]interface{}) {
	session.mu.Lock()

	if playing, ok := msg["playing"].(bool); ok {
		session.Playing = playing
	}

	if currentTime, ok := msg["currentTime"].(float64); ok {
		session.CurrentTime = currentTime
	}

	session.LastUpdate = time.Now()
	session.mu.Unlock()

	session.BroadcastSync(peer.ID)

	log.Printf("Sync from peer %s: playing=%v, time=%.2f",
		peer.ID, session.Playing, session.CurrentTime)
}

// handleVideoDataMessage обрабатывает загрузку видео
func handleVideoDataMessage(peer *Peer, session *Session, storage *VideoStorage, msg map[string]interface{}) {
	var videoData []byte

	if dataStr, ok := msg["data"].(string); ok {
		// Если это base64 строка, декодируем
		decoded, err := base64.StdEncoding.DecodeString(dataStr)
		if err != nil {
			log.Printf("Failed to decode base64 from peer %s: %v", peer.ID, err)
			return
		}
		videoData = decoded
	} else if dataArr, ok := msg["data"].([]interface{}); ok {
		// Если это массив байтов
		videoData = make([]byte, len(dataArr))
		for i, v := range dataArr {
			if b, ok := v.(float64); ok {
				videoData[i] = byte(b)
			}
		}
	}

	if len(videoData) == 0 {
		log.Printf("Empty video data from peer %s", peer.ID)
		return
	}

	storage.StoreVideo(session.ID, videoData)
	log.Printf("Video stored for session %s (size: %d bytes)", session.ID, len(videoData))

	session.mu.RLock()
	peers := session.Peers
	session.mu.RUnlock()

	for id, p := range peers {
		if id != peer.ID {
			p.Send(map[string]interface{}{
				"type": "videoData",
				"data": base64.StdEncoding.EncodeToString(videoData),
			})
		}
	}

	log.Printf("Video broadcast to %d peers", len(peers)-1)
}
