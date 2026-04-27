package hub

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
	"go104/internal/models"
)

// DatapointMsg is broadcast when a new value arrives from a slave.
type DatapointMsg struct {
	Type       string  `json:"type"`
	SignalID   int64   `json:"signal_id"`
	LineID     int64   `json:"line_id"`
	IOA        int     `json:"ioa"`
	Name       string  `json:"name"`
	SignalType string  `json:"signal_type"`
	TypeID     int     `json:"type_id"`
	Unit       string  `json:"unit"`
	Value      float64 `json:"value"`
	RawValue   float64 `json:"raw_value"`
	Quality    uint8   `json:"quality"`
	QualityOK  bool    `json:"quality_ok"`
	Timestamp  string  `json:"timestamp"`
	ReceivedAt string  `json:"received_at"`
}

// LineStatusMsg is broadcast on every connection state change.
type LineStatusMsg struct {
	Type    string `json:"type"`
	LineID  int64  `json:"line_id"`
	State   string `json:"state"`
	RxCount int64  `json:"rx_count"`
	TxCount int64  `json:"tx_count"`
	Addr    string `json:"addr"`
}

// Client represents one connected WebSocket peer.
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

// Hub manages all WebSocket clients and routes broadcasts.
type Hub struct {
	clients    map[*Client]struct{}
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func New() *Hub {
	return &Hub{
		clients:    make(map[*Client]struct{}),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run must be called in a dedicated goroutine.
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.clients[c] = struct{}{}
		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.send)
			}
		case msg := <-h.broadcast:
			for c := range h.clients {
				select {
				case c.send <- msg:
				default:
					delete(h.clients, c)
					close(c.send)
				}
			}
		}
	}
}

// BroadcastDatapoint serialises a datapoint and sends to all clients.
func (h *Hub) BroadcastDatapoint(dp *models.Datapoint) {
	msg := DatapointMsg{
		Type:       "datapoint",
		SignalID:   dp.SignalID,
		LineID:     dp.LineID,
		IOA:        dp.IOA,
		Name:       dp.Name,
		SignalType: string(dp.Kind),
		TypeID:     dp.TypeID,
		Unit:       dp.Unit,
		Value:      dp.Value,
		RawValue:   dp.RawValue,
		Quality:    dp.Quality,
		QualityOK:  dp.QualityOK(),
		Timestamp:  dp.Timestamp.UTC().Format(time.RFC3339Nano),
		ReceivedAt: dp.ReceivedAt.UTC().Format(time.RFC3339Nano),
	}
	h.broadcastJSON(msg)
}

// BroadcastLineStatus sends a line state update to all clients.
func (h *Hub) BroadcastLineStatus(lineID int64, state string, rx, tx int64, addr string) {
	h.broadcastJSON(LineStatusMsg{
		Type:    "line_status",
		LineID:  lineID,
		State:   state,
		RxCount: rx,
		TxCount: tx,
		Addr:    addr,
	})
}

// BroadcastRaw sends an arbitrary JSON-marshallable value.
func (h *Hub) BroadcastRaw(v any) {
	h.broadcastJSON(v)
}

func (h *Hub) broadcastJSON(v any) {
	data, err := json.Marshal(v)
	if err != nil {
		slog.Error("hub marshal", "err", err)
		return
	}
	select {
	case h.broadcast <- data:
	default:
		slog.Warn("hub broadcast channel full, dropping message")
	}
}

// ServeWS upgrades an HTTP connection and pumps messages.
func (h *Hub) ServeWS(conn *websocket.Conn) {
	c := &Client{hub: h, conn: conn, send: make(chan []byte, 128)}
	h.register <- c

	// writer goroutine
	go func() {
		defer func() {
			h.unregister <- c
			conn.Close()
		}()
		for {
			msg, ok := <-c.send
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, nil) //nolint:errcheck
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		}
	}()

	// reader loop — keeps connection alive, handles pings
	conn.SetReadLimit(512)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second)) //nolint:errcheck
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second)) //nolint:errcheck
		return nil
	})
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
