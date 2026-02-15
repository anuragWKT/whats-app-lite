package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"whats-app-lite/internal/client"
	"whats-app-lite/internal/models"
	"whats-app-lite/internal/server"
	"whats-app-lite/internal/storage"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var userManager *storage.UserManager

func serveWs(hub *server.Hub, w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")

	if username == "" || password == "" {
		http.Error(w, "Missing username or password", http.StatusBadRequest)
		return
	}

	if err := userManager.Authenticate(username, password); err != nil {
		http.Error(w, "Invalid credentials", http.StatusForbidden)
		return
	}

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
	c := client.NewClient(conn, room.Broadcast, room.Unregister, username)
	room.Register <- c
	go c.WritePump()
	go c.ReadPump()
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.User
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password required", http.StatusBadRequest)
		return
	}

	if err := userManager.Register(req.Username, req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "User registered successfully",
		"username": req.Username,
	})
}

func handleRooms(hub *server.Hub, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")

	if username == "" || password == "" {
		http.Error(w, "Missing username or password", http.StatusBadRequest)
		return
	}

	if err := userManager.Authenticate(username, password); err != nil {
		http.Error(w, "Invalid credentials", http.StatusForbidden)
		return
	}

	activeRooms := hub.ListRooms()
	fileManager := storage.NewFileManager("./storage_data")
	persistedRooms, err := fileManager.ListPersistedRooms()
	if err != nil {
		http.Error(w, "Failed to load rooms", http.StatusInternalServerError)
		return
	}

	roomSet := make(map[string]bool)
	for _, room := range activeRooms {
		roomSet[room] = true
	}
	for _, room := range persistedRooms {
		roomSet[room] = true
	}

	rooms := make([]string, 0, len(roomSet))
	for room := range roomSet {
		rooms = append(rooms, room)
	}
	sort.Strings(rooms)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{
		"rooms": rooms,
	})
}

func main() {
	userManager = storage.NewUserManager("./storage_data/users.json")

	hub := server.NewHub()

	mux := http.NewServeMux()

	mux.HandleFunc("/register", handleRegister)
	mux.HandleFunc("/rooms", func(w http.ResponseWriter, r *http.Request) {
		handleRooms(hub, w, r)
	})

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	fs := http.FileServer(http.Dir("./web"))
	mux.Handle("/", fs)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		log.Println("Server started on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited cleanly")
}