package room

import (
	"encoding/json"
	"log"
	"whats-app-lite/internal/client"
	"whats-app-lite/internal/models"
	"whats-app-lite/internal/storage"
)

type Room struct {
	Name       string
	clients    map[*client.Client]bool
	Broadcast  chan []byte
	Register   chan *client.Client
	Unregister chan *client.Client
	Storage    *storage.FileManager
}

func NewRoom(name string) *Room {
	return &Room{
		Name:       name,
		Broadcast:  make(chan []byte),
		Register:   make(chan *client.Client),
		Unregister: make(chan *client.Client),
		clients:    make(map[*client.Client]bool),
		Storage:    storage.NewFileManager("./storage_data"),
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

			clientActive := true
			for _, msgBytes := range history {
				if !clientActive {
					break
				}

				var msg models.Message
				if err := json.Unmarshal(msgBytes, &msg); err != nil {
					continue
				}

				isPublic := msg.Recipient == ""
				isForMe := msg.Recipient == client.Username
				isFromMe := msg.Sender == client.Username

				if isPublic || isForMe || isFromMe {
					select {
					case client.Send <- msgBytes:
					default:
						close(client.Send)
						delete(r.clients, client)
						clientActive = false
					}
				}
			}

		case client := <-r.Unregister:
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.Send)
			}

		case message := <-r.Broadcast:
			var msg models.Message
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Invalid JSON received: %v", err)
				continue
			}

			if msg.Type != "typing" {
				if err := r.Storage.SaveMessage(r.Name, message); err != nil {
					log.Printf("Error saving message: %v", err)
				}
			}

			if msg.Recipient != "" {
				for client := range r.clients {
					if client.Username == msg.Recipient || client.Username == msg.Sender {
						select {
						case client.Send <- message:
						default:
							close(client.Send)
							delete(r.clients, client)
						}
					}
				}
			} else {
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
}
