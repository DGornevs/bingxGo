package websocket

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// ====== DATA STRUCTURES ======

type PriceUpdate struct {
	Type   string  `json:"type"`
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
}

type Channel struct {
	ID       string `json:"id"`
	ReqType  string `json:"reqType"`
	DataType string `json:"dataType"`
}

type MarketData struct {
	DataType string `json:"dataType"`
	Data     struct {
		Price string `json:"p"`
	} `json:"data"`
}

// ====== MAIN STRUCT ======

type BingXWebSocket struct {
	path           string
	conn           *websocket.Conn
	tokens         []string
	messageHandler func(PriceUpdate)
	mu             sync.RWMutex
	quit           chan struct{}
}

// ====== CONSTRUCTOR ======

func NewBingXWebSocket(tokens []string, messageHandler func(PriceUpdate)) *BingXWebSocket {
	return &BingXWebSocket{
		path:           "wss://open-api-swap.bingx.com/swap-market",
		tokens:         tokens,
		messageHandler: messageHandler,
		quit:           make(chan struct{}),
	}
}

// ====== CONNECTION ======

func (ws *BingXWebSocket) Connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(ws.path, nil)
	if err != nil {
		return fmt.Errorf("WebSocket connection failed: %w", err)
	}

	ws.mu.Lock()
	ws.conn = conn
	ws.mu.Unlock()

	log.Println("✅ WebSocket connected")

	go ws.listen()
	return ws.subscribeAll()
}

// subscribeAll subscribes to all tokens with a short delay to avoid flooding
func (ws *BingXWebSocket) subscribeAll() error {
	for i, token := range ws.tokens {
		channel := Channel{
			ID:       uuid.New().String(),
			ReqType:  "sub",
			DataType: token,
		}

		data, err := json.Marshal(channel)
		if err != nil {
			return fmt.Errorf("marshal channel %s: %w", token, err)
		}

		ws.mu.RLock()
		conn := ws.conn
		ws.mu.RUnlock()

		if conn == nil {
			return fmt.Errorf("WebSocket not connected")
		}

		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return fmt.Errorf("subscribe %s: %w", token, err)
		}

		log.Printf("Subscribed to %s (%d/%d)", token, i+1, len(ws.tokens))
		if i < len(ws.tokens)-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}
	return nil
}

// ====== MESSAGE LOOP ======

func (ws *BingXWebSocket) listen() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered in listener: %v", r)
		}
	}()

	for {
		ws.mu.RLock()
		conn := ws.conn
		ws.mu.RUnlock()

		if conn == nil {
			return
		}

		select {
		case <-ws.quit:
			return
		default:
		}

		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			ws.reconnect()
			return
		}

		ws.handleMessage(msg)
	}
}

// ====== MESSAGE HANDLING ======

func (ws *BingXWebSocket) handleMessage(msg []byte) {
	data, err := decompress(msg)
	if err != nil {
		log.Printf("Decompression failed: %v", err)
		return
	}

	if bytes.Equal(data, []byte("Ping")) {
		ws.sendPong()
		return
	}

	var market MarketData
	if err := json.Unmarshal(data, &market); err != nil {
		log.Printf("JSON decode failed: %v", err)
		return
	}

	ws.dispatch(market)
}

func (ws *BingXWebSocket) dispatch(m MarketData) {
	if m.DataType == "" || m.Data.Price == "" {
		return
	}

	symbol := parseSymbol(m.DataType)
	price, err := parsePrice(m.Data.Price)
	if err != nil {
		log.Printf("Invalid price %s: %v", m.Data.Price, err)
		return
	}

	if ws.messageHandler != nil {
		ws.messageHandler(PriceUpdate{
			Type:   "priceUpdate",
			Symbol: symbol,
			Price:  price,
		})
	}
}

// ====== UTILITIES ======

func decompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

func parseSymbol(s string) string {
	if idx := strings.IndexRune(s, '@'); idx > 0 {
		return s[:idx]
	}
	return s
}

func parsePrice(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

func (ws *BingXWebSocket) sendPong() {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	if ws.conn != nil {
		_ = ws.conn.WriteMessage(websocket.TextMessage, []byte("Pong"))
	}
}

// ====== RECONNECT & CLOSE ======

func (ws *BingXWebSocket) reconnect() {
	log.Println("Attempting WebSocket reconnection...")
	ws.Close()
	time.Sleep(2 * time.Second)
	if err := ws.Connect(); err != nil {
		log.Printf("Reconnection failed: %v", err)
	} else {
		log.Println("Reconnected successfully")
	}
}

func (ws *BingXWebSocket) Close() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	select {
	case <-ws.quit:
	default:
		close(ws.quit)
	}

	if ws.conn != nil {
		err := ws.conn.Close()
		ws.conn = nil
		if err != nil {
			return fmt.Errorf("error closing WebSocket: %w", err)
		}
		log.Println("WebSocket closed")
	}
	return nil
}
