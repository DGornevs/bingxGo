package main

import (
	"fmt"
	"log"

	"bingxGo/config"
	"bingxGo/internal/binance"
	"bingxGo/internal/bingx"
	"bingxGo/internal/parser"
	"bingxGo/internal/telegram"
)

func main() {
	client := binance.NewClient()

	pairs, err := client.FetchPairs()
	if err != nil {
		log.Fatalf("Error fetching Binance pairs: %v", err)
	}

	fmt.Printf("Fetched %d USDT pairs:\n", len(pairs))
	for i, p := range pairs {
		fmt.Printf("%2d. %s\n", i+1, p.Symbol)
	}
	fmt.Println("===================================")

	pairsBingX, err := bingx.FetchPairs()
	if err != nil {
		log.Fatalf("Error fetching BingX pairs: %v", err)
	}

	fmt.Printf("Fetched %d BingX pairs:\n", len(pairsBingX))
	for i, p := range pairsBingX {
		fmt.Printf("%2d. %s\n", i+1, p)
	}

	fmt.Println("===================================")
	cfg := config.Load()
	if cfg == nil {
		log.Fatalf("Error loading config")
	}
	fmt.Printf("Config loaded: %+v\n", cfg)

	tg := telegram.New(cfg.CHAT_BOT_TOKEN)
	err = tg.SendMessage(cfg.CHAT_ID, "<b>Hello!</b> This is a test message.")
	if err != nil {
		log.Printf("Error sending Telegram message: %v", err)
	} else {
		log.Println("Telegram message sent successfully")
	}

	titles := []string{
		"Binance Will Delist BTCUSDT, ETHUSDT, XRPUSDT on 2025-10-31",
		"Binance Announced the First Batch of Vote to Delist Results and Will Delist SOLUSDT, ADAUSDT on 2025-12-01",
		"Random Non-Matching Title",
	}

	for _, t := range titles {
		pairs := parser.ExtractPairs(t)
		fmt.Printf("Title: %q → Pairs: %v\n", t, pairs)
	}
}
