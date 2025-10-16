package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Telegram represents a Telegram bot client
type Telegram struct {
	apiURL   string
	botToken string
	client   *http.Client
}

// TelegramResponse represents the response from Telegram API
type TelegramResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description,omitempty"`
	Result      struct {
		MessageID int `json:"message_id"`
	} `json:"result,omitempty"`
}

// MessagePayload represents the payload for sending a message
type MessagePayload struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

// New creates a new Telegram client using a bot token.
func New(botToken string) *Telegram {
	return &Telegram{
		apiURL:   "https://api.telegram.org",
		botToken: botToken,
		client:   &http.Client{Timeout: 15 * time.Second},
	}
}

// SendMessage sends a message to a specific Telegram chat ID.
func (t *Telegram) SendMessage(chatID, message string) error {
	url := fmt.Sprintf("%s/bot%s/sendMessage", t.apiURL, t.botToken)

	payload := MessagePayload{
		ChatID:    chatID,
		Text:      message,
		ParseMode: "HTML",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	var tgResp TelegramResponse
	if err := json.NewDecoder(resp.Body).Decode(&tgResp); err != nil {
		return fmt.Errorf("decode telegram response: %w", err)
	}

	if !tgResp.OK {
		return fmt.Errorf("telegram API error: %s", tgResp.Description)
	}

	return nil
}
