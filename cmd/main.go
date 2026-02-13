package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"whats-app-lite/internal/client"
	"whats-app-lite/internal/server"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func serveWs(hub *server.Hub, w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	roomName := r.URL.Query().Get("room")
	if roomName == "" {
		roomName = "general"
	}

	room := hub.GetOrCreateRoom(roomName)
	c := client.NewClient(conn, room.Broadcast, room.Unregister)
	room.Register <- c

	go c.WritePump()
	go c.ReadPump()
}

func main() {

	hub := server.NewHub()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	log.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
