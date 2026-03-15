package main

import (
	"log"
	"net/http"

	"backend/handlers"
)

func main() {

	// Start the broadcast hub
	go handlers.HandleBroadcast()

	// WebSocket endpoint
	http.HandleFunc("/ws", handlers.HandleWebSockets)

	// Optional health check route
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Whiteboard server running"))
	})

	log.Println("Server started on :8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}