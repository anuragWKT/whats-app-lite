package room

import (
	"whats-app-lite/internal/client"

	"golang.org/x/text/message"
)

type Room struct{
	Name string
	clients map[*client.Client]bool
	Broadcast chan []byte
	Register chan *client.Client
	Unregister chan *client.Client
}

func NewRoom(name string)*Room{
	return &Room{
		Name: name,
		Broadcast: make(chan []byte),
		Register: make(chan *client.Client),
		Unregister: make(chan *client.Client),
		clients: make(map[*client.Client]bool),
	}
}

func (r *Room) Run(){
	for{
		select {
		case client:=<-r.Register:
			r.clients[client]=true
		
		case client:=<-r.Unregister:
			if _,ok:=r.clients[client];ok{
				delete(r.clients,client)
				close(client.Send)
			}
		case message:=<-r.Broadcast:
			for client:=range r.clients{
				select{
				case client.Send<-message:
				default:
					close(client.Send)
					delete(r.clients,client)
				}
			}
		}
	}
}