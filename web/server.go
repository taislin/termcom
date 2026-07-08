package web

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
	Cols int    `json:"cols,omitempty"`
	Rows int    `json:"rows,omitempty"`
}

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

type Server struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func NewServer() *Server {
	return &Server{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (s *Server) Run() {
	for {
		select {
		case client := <-s.register:
			s.clients[client] = true
			log.Printf("Client connected. Total: %d", len(s.clients))

		case client := <-s.unregister:
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.send)
				log.Printf("Client disconnected. Total: %d", len(s.clients))
			}

		case message := <-s.broadcast:
			for client := range s.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(s.clients, client)
				}
			}
		}
	}
}

func (s *Server) Broadcast(data []byte) {
	s.broadcast <- data
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
	}

	s.register <- client

	go client.writePump()
	go client.readPump(s)
}

func (c *Client) readPump(s *Server) {
	defer func() {
		s.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		// Handle different message types
		switch msg.Type {
		case "input":
			// Forward input to game
			if InputHandler != nil {
				InputHandler(msg.Data)
			}
		case "resize":
			// Handle resize
			if ResizeHandler != nil {
				ResizeHandler(msg.Cols, msg.Rows)
			}
		}
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()

	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			break
		}
	}
}

// Game handlers - will be set by game
var InputHandler func(input string)
var ResizeHandler func(cols, rows int)

func (s *Server) SendOutput(output string) {
	msg := Message{
		Type: "output",
		Data: output,
	}
	data, _ := json.Marshal(msg)
	s.Broadcast(data)
}

func StartServer(addr string) *Server {
	server := NewServer()
	go server.Run()

	http.HandleFunc("/ws", server.HandleWebSocket)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/index.html")
	})

	log.Printf("Web server starting on %s", addr)
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatal(err)
		}
	}()

	return server
}
