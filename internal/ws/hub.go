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
	RoomID string `json:"room_id,omitempty"`
}

type Room struct {
	ID         string
	HostID     string
	clients    map[*Client]bool
	mu         sync.RWMutex
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
}

func NewRoom(id string, hostId string) *Room {
	return &Room{
		ID:         id,
		HostID:     hostId,
		clients:    make(map[*Client]bool),
		broadcast:  make(chan Message, 1024),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
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

	//rooms
	rooms map[string]*Room

	// Worker pool settings
	workerCount int
	jobQueue    chan Message
}

// for the rooms
// NewHub initializes a Hub with all required channels and maps.
func NewHub(workerCount int) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		broadcast:   make(chan Message),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		rooms:       make(map[string]*Room),
		workerCount: workerCount,
		// Buffered jobQueue so broadcasting doesn't block writers unnecessarily
		jobQueue: make(chan Message, workerCount*10),
	}
}

// create the Room by id its a thread-safe operation
func (h *Hub) CreateRoom(id string, hostId string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.createRoom(id, hostId)
}

// search the room by id its a thread-safe operation
func (h *Hub) GetRoom(id string) *Room {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.rooms[id]
}

// remove the room by id its a lock free operation
func (h *Hub) RemoveRoom(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.rooms, id)
}

// get all the rooms its a thread-safe operation.
// but will not be used anywhere
func (h *Hub) GetAllRooms() []*Room {
	h.mu.RLock()
	defer h.mu.RUnlock()
	rooms := make([]*Room, 0, len(h.rooms))
	for _, room := range h.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

// getRoom internal method (non-locking)
func (h *Hub) getRoom(id string) *Room {
	return h.rooms[id]
}

// createRoom internal method (non-locking)
func (h *Hub) createRoom(id string, hostId string) *Room {
	room := NewRoom(id, hostId)
	h.rooms[id] = room
	return room
}

// register the client to the room its a thread-safe operation
func (h *Hub) RegisterClientToRoom(client *Client, roomId string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	room := h.getRoom(roomId)
	if room == nil {
		return
	}
	room.mu.Lock()
	defer room.mu.Unlock()
	room.clients[client] = true
}

// unregister the client from the room its a thread-safe operation
func (h *Hub) UnregisterClientFromRoom(client *Client, roomId string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	room := h.getRoom(roomId)
	if room == nil {
		return
	}
	room.mu.Lock()
	defer room.mu.Unlock()
	delete(room.clients, client)
}

// broadcast the message to the room its a lock free operation
func (h *Hub) BroadcastToRoom(roomId string, message Message) {
	room := h.GetRoom(roomId)
	if room == nil {
		return
	}
	room.mu.Lock()
	defer room.mu.Unlock()
	room.broadcast <- message
}

// run the hub its a lock free operation
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
		// If RoomID is provided, only broadcast to clients in that room
		if message.RoomID != "" && client.RoomID != message.RoomID {
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
