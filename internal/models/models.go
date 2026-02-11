package models
//the single piece of data exchanged between the client and the server
type Message struct{
	Type string `json:"type"`//indicates the action like join,leave,message,typing,private
	Sender string `json:"sender"`
	Recipient string `json:"recipient,omitempty"`
	Content string `json:"content"`
	Room string `json:"room"`
}
//represents the registered user for the authentication
type User struct{
	Username string `json:"username"`
	Password string `json:"password"`
}