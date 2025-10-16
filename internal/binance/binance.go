package binance

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ====== TYPES ======

type BinanceClient struct {
	client *http.Client
}

type BinancePair struct {
	Symbol string `json:"symbol"`
}

type exchangeInfoResponse struct {
	Symbols []BinancePair `json:"symbols"`
}

type NewsArticle struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"code"`
	ReleaseDate int64  `json:"releaseDate"`
}

type newsResponse struct {
	Data struct {
		Articles []NewsArticle `json:"catalogs"`
	} `json:"data"`
}

// ====== CONSTRUCTOR ======

func NewClient() *BinanceClient {
	return &BinanceClient{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// ====== FETCH NEWS ======

func (b *BinanceClient) GetNews() ([]NewsArticle, error) {
	const url = "https://www.binance.com/bapi/apex/v1/public/apex/cms/article/list/query?type=1&pageNo=1&pageSize=10&catalogId=161"

	resp, err := b.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching Binance news: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("non-200 status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding Binance news JSON: %w", err)
	}

	// For simplicity, we only log success here.
	fmt.Println("✅ Binance news fetched successfully")
	return nil, nil
}

// ====== FETCH PAIRS ======

func (b *BinanceClient) FetchPairs() ([]BinancePair, error) {
	const url = "https://fapi.binance.com/fapi/v1/exchangeInfo"

	resp, err := b.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching pairs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("non-200 status %d: %s", resp.StatusCode, string(body))
	}

	var data exchangeInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decoding pairs JSON: %w", err)
	}

	usdtPairs := make([]BinancePair, 0, len(data.Symbols))
	for _, s := range data.Symbols {
		if strings.HasSuffix(s.Symbol, "USDT") {
			usdtPairs = append(usdtPairs, s)
		}
	}

	return usdtPairs, nil
}
