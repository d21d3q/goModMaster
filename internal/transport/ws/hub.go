package ws

import (
	"context"
	"net/http"
	"sync"

	"gomodmaster/internal/core"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type Hub struct {
	clients    map[*websocket.Conn]bool
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	done       chan struct{}
	doneOnce   sync.Once
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		done:       make(chan struct{}),
	}
}

func (h *Hub) Run(ctx context.Context, events <-chan core.Event) {
	defer h.closeDone()
	for {
		select {
		case <-ctx.Done():
			h.closeAll()
			return
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				_ = client.Close()
			}
		case event := <-events:
			for client := range h.clients {
				_ = client.WriteJSON(event)
			}
		}
	}
}

func (h *Hub) Handle(c echo.Context, initialEvents ...core.Event) error {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	for _, event := range initialEvents {
		if err := conn.WriteJSON(event); err != nil {
			_ = conn.Close()
			return err
		}
	}

	if !h.registerConn(conn) {
		_ = conn.Close()
		return nil
	}
	go func() {
		defer func() { h.unregisterConn(conn) }()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	return nil
}

func (h *Hub) closeAll() {
	for client := range h.clients {
		_ = client.Close()
		delete(h.clients, client)
	}
}

func (h *Hub) closeDone() {
	h.doneOnce.Do(func() {
		close(h.done)
	})
}

func (h *Hub) registerConn(conn *websocket.Conn) bool {
	select {
	case h.register <- conn:
		return true
	case <-h.done:
		return false
	}
}

func (h *Hub) unregisterConn(conn *websocket.Conn) {
	select {
	case h.unregister <- conn:
	case <-h.done:
		_ = conn.Close()
	}
}
