package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	DefaultPort = "8080"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	storage := NewVideoStorage()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(w, r, storage)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	// Путь для production
	frontendPath := "./frontend"
	if _, err := os.Stat("../frontend"); err == nil {
		frontendPath = "../frontend"
	}

	http.Handle("/", http.FileServer(http.Dir(frontendPath)))

	addr := ":" + port
	log.Printf("Server starting on port %s", port)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
