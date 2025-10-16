package bingx

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ====== CONSTANTS ======
const (
	baseURL = "https://open-api.bingx.com"
	timeout = 30 * time.Second
)

// ====== HTTP HELPERS ======

// doRequest simplifies making API requests with automatic JSON decoding
func doRequest(req *http.Request, v interface{}) error {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%s request failed: %v", ts(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s HTTP %d: %s", ts(), resp.StatusCode, body)
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("%s JSON decode error: %v", ts(), err)
	}
	return nil
}

// buildQuery builds a sorted and URL-encoded query string
func buildQuery(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, url.QueryEscape(params[k])))
	}
	return strings.Join(parts, "&")
}

// signQuery generates an HMAC SHA256 signature
func signQuery(secret, query string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(query))
	return hex.EncodeToString(h.Sum(nil))
}

// ts returns formatted timestamp string
func ts() string { return time.Now().Format("2006-01-02 15:04:05") }

// ====== BASIC ENDPOINTS ======

func KeepAlive() {
	resp, err := http.Get(baseURL + "/openApi/spot/v1/server/time")
	if err != nil {
		fmt.Printf("Error fetching server time: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Decode error: %v\n", err)
		return
	}
	fmt.Printf("Server time: %v\n", result["serverTime"])
}

func FetchPairs() ([]string, error) {
	resp, err := http.Get(baseURL + "/openApi/swap/v2/quote/contracts")
	if err != nil {
		return nil, fmt.Errorf("fetch pairs: %v", err)
	}
	defer resp.Body.Close()

	var r struct {
		Data []struct{ Symbol string } `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("decode pairs: %v", err)
	}

	pairs := make([]string, len(r.Data))
	for i, d := range r.Data {
		pairs[i] = d.Symbol
	}
	return pairs, nil
}

// ====== STRUCTS ======
type WalletBalanceResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Balance []BalanceItem `json:"balance"`
	} `json:"data"`
}

type BalanceItem struct {
	Asset              string `json:"asset"`
	Balance            string `json:"balance"`
	CrossWalletBalance string `json:"crossWalletBalance"`
	CrossUnPnl         string `json:"crossUnPnl"`
	AvailableBalance   string `json:"availableBalance"`
	MaxWithdrawAmount  string `json:"maxWithdrawAmount"`
	MarginAvailable    bool   `json:"marginAvailable"`
	UpdateTime         int64  `json:"updateTime"`
}

type PricesResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data []PriceItem `json:"data"`
}

type PriceItem struct {
	Symbol, MarkPrice, IndexPrice, EstimatedSettlePrice, LastFundingRate, InterestRate string
	NextFundingTime, Time                                                              int64
}

type LeverageResponse struct {
	Code int          `json:"code"`
	Msg  string       `json:"msg"`
	Data LeverageData `json:"data"`
}

type LeverageData struct {
	Symbol, LongLeverage, ShortLeverage, MaxLongLeverage, MaxShortLeverage string
}

type BatchTradeResponse struct {
	Code int                `json:"code"`
	Msg  string             `json:"msg"`
	Data []BatchTradeResult `json:"data"`
}

type BatchTradeResult struct {
	OrderID, ClientOrderID, Symbol, Status, Error string
}

type BatchOrder struct {
	Symbol, Side, PositionSide, Type, Quantity, Price, ClientOrderID, TimeInForce string
}

// ====== API CALLS ======

func GetWalletBalance(apiKey, apiSecret string) (*WalletBalanceResponse, error) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	query := "timestamp=" + timestamp
	signature := signQuery(apiSecret, query)

	url := fmt.Sprintf("%s/openApi/swap/v3/user/balance?%s&signature=%s", baseURL, query, signature)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-BX-APIKEY", apiKey)

	var res WalletBalanceResponse
	if err := doRequest(req, &res); err != nil {
		return nil, err
	}
	if res.Code != 0 {
		return nil, fmt.Errorf("%s API error: %s", ts(), res.Msg)
	}
	return &res, nil
}

func FetchPrices() (*PricesResponse, error) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	url := fmt.Sprintf("%s/openApi/swap/v2/quote/premiumIndex?timestamp=%s", baseURL, timestamp)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-BX-TIMESTAMP", timestamp)

	var res PricesResponse
	if err := doRequest(req, &res); err != nil {
		return nil, err
	}
	if res.Code != 0 {
		return nil, fmt.Errorf("%s API error: %s", ts(), res.Msg)
	}
	return &res, nil
}

func FetchLeverage(apiKey, apiSecret, symbol string) (*LeverageResponse, error) {
	params := map[string]string{
		"symbol":    symbol,
		"timestamp": strconv.FormatInt(time.Now().UnixMilli(), 10),
	}
	query := buildQuery(params)
	signature := signQuery(apiSecret, query)

	url := fmt.Sprintf("%s/openApi/swap/v2/trade/leverage?%s&signature=%s", baseURL, query, signature)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-BX-APIKEY", apiKey)

	var res LeverageResponse
	if err := doRequest(req, &res); err != nil {
		return nil, err
	}
	if res.Code != 0 {
		return nil, fmt.Errorf("%s API error: %s", ts(), res.Msg)
	}
	return &res, nil
}

func BatchTrade(apiKey, apiSecret string, orders []BatchOrder) (*BatchTradeResponse, error) {
	body, err := json.Marshal(orders)
	if err != nil {
		return nil, fmt.Errorf("marshal orders: %v", err)
	}

	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	unsigned := fmt.Sprintf("batchOrders=%s&timestamp=%s", body, timestamp)
	signature := signQuery(apiSecret, unsigned)

	url := fmt.Sprintf("%s/openApi/swap/v2/trade/batchOrders?batchOrders=%s&timestamp=%s&signature=%s",
		baseURL, url.QueryEscape(string(body)), timestamp, signature)

	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("X-BX-APIKEY", apiKey)

	var res BatchTradeResponse
	if err := doRequest(req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
