package room

import (
	"log"
	"whats-app-lite/internal/client"
	"whats-app-lite/internal/storage"
)

type Room struct {
	Name string
	clients map[*client.Client]bool
	Broadcast chan []byte
	Register chan *client.Client
	Unregister chan *client.Client
	Storage *storage.FileManager
}

func NewRoom(name string) *Room {
	return &Room{
		Name:       name,
		Broadcast:  make(chan []byte),
		Register:   make(chan *client.Client),
		Unregister: make(chan *client.Client),
		clients:    make(map[*client.Client]bool),
		Storage: storage.NewFileManager("./storage_data"),
	}
}

func (r *Room) Run() {
	for {
		select {

		case client := <-r.Register:
			r.clients[client] = true

			history, err := r.Storage.LoadHistory(r.Name)
			if err != nil {
				log.Printf("Error loading history for room %s: %v", r.Name, err)
			}

			for _, msg := range history {
				select {
				case client.Send <- msg:

				default:

					close(client.Send)
					delete(r.clients, client)
					break 
				}
			}

		case client := <-r.Unregister:
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.Send)
			}

		case message := <-r.Broadcast:
			if err := r.Storage.SaveMessage(r.Name, message); err != nil {
				log.Printf("Error saving message in room %s: %v", r.Name, err)
			}

			for client := range r.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(r.clients, client)
				}
			}
		}
	}
}