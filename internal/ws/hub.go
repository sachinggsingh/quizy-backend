package ws

import (
	"encoding/json"
	"log"
	"sync"
)

type Message struct {
	Type   string `json:"type"`
	Data   any    `json:"data"`
	QuizID string `json:"quiz_id,omitempty"`
}

type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan Message

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	mu sync.RWMutex

	// Worker pool settings
	workerCount int
	jobQueue    chan Message
}

func NewHub(workerCount int) *Hub {
	return &Hub{
		broadcast:   make(chan Message, 1024),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		clients:     make(map[*Client]bool),
		jobQueue:    make(chan Message, 2048),
		workerCount: workerCount,
	}
}

func (h *Hub) Run() {
	// Start workers
	for i := 0; i < h.workerCount; i++ {
		go h.worker()
	}

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client registered. Total clients: %d", len(h.clients))
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("Client unregistered. Total clients: %d", len(h.clients))
		case message := <-h.broadcast:
			// Push message to worker pool
			h.jobQueue <- message
		}
	}
}

func (h *Hub) worker() {
	for message := range h.jobQueue {
		h.broadcastToClients(message)
	}
}

func (h *Hub) broadcastToClients(message Message) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling message: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		// If QuizID is provided, only broadcast to clients in that quiz
		if message.QuizID != "" && client.QuizID != message.QuizID {
			continue
		}

		select {
		case client.Send <- data:
		default:
			// If buffer is full, unregister client
			// This is handled in the next loop of Run or via writePump error
			go func(c *Client) {
				h.unregister <- c
			}(client)
		}
	}
}

func (h *Hub) Broadcast(msg Message) {
	h.broadcast <- msg
}
