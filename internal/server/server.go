package server

import (
	"sort"
	"sync"
	"whats-app-lite/internal/room"
)

type Hub struct {
	rooms map[string]*room.Room
	mu    sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]*room.Room),
	}
}

func (h *Hub) GetOrCreateRoom(name string) *room.Room {
	h.mu.RLock()
	if r, ok := h.rooms[name]; ok {
		h.mu.RUnlock()
		return r
	}
	h.mu.RUnlock()
	h.mu.Lock()
	defer h.mu.Unlock()

	if r, ok := h.rooms[name]; ok {
		return r
	}
	newRoom := room.NewRoom(name)
	go newRoom.Run()

	h.rooms[name] = newRoom
	return newRoom
}

func (h *Hub) ListRooms() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	rooms := make([]string, 0, len(h.rooms))
	for name := range h.rooms {
		rooms = append(rooms, name)
	}
	sort.Strings(rooms)
	return rooms
}
