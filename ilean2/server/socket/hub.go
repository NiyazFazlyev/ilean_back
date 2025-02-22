package socket

import (
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

			logrus.Info(len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

		case message := <-h.broadcast:

			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *Hub) Send(typeClient string, serialNumber int, message []byte) {

	//logrus.Info("send ")
	//
	//inc, err := json.Marshal(incidents)
	//
	//if err != nil {
	//	logrus.Info("failed to marshal")
	//}

	for client := range h.clients {

		if client.typeClient == typeClient && client.serialNumber == serialNumber {
			err := client.conn.WriteMessage(websocket.TextMessage, message)

			if err != nil {
				logrus.Error(err)
			}

			logrus.WithField(typeClient, serialNumber).Info("send messages")

			//	client.conn.SetWriteDeadline(time.Now().Add(writeWait))
		}

		//select {
		//case client.send <- []byte(s) :
		//	logrus.Info("succes send ")
		//default:
		//	close(client.send)
		//	delete(h.clients, client)
		//}
	}
}

func (h *Hub) GetCountClient() int {

	return len(h.clients)
}
