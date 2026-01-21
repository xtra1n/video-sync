package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const (
	DefaultPort = "8080"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	storage := NewVideoStorage()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(w, r, storage)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	// Находим frontend папку
	ex, _ := os.Executable()
	frontendPath := filepath.Join(filepath.Dir(ex), "..", "frontend")

	// Проверяем где мы запускаемся
	if _, err := os.Stat("../frontend"); err == nil {
		frontendPath = "../frontend"
	}
	if _, err := os.Stat("./frontend"); err == nil {
		frontendPath = "./frontend"
	}

	log.Printf("Serving frontend from: %s", frontendPath)
	http.Handle("/", http.FileServer(http.Dir(frontendPath)))

	addr := ":" + port
	log.Printf("Server starting on http://localhost%s", addr)
	log.Printf("Open http://localhost%s in browser", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
