package chaindataagg

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/sync/errgroup"
)

type Token struct {
	ID        string            `json:"id"`
	Symbol    string            `json:"symbol"`
	Name      string            `json:"name"`
	Platforms map[string]string `json:"platforms"`
}

type TokenPrice struct {
	TokenSymbol  string
	Date         string
	AveragePrice float64
}

// FetchTokenList fetches the list of tokens from CoinGecko.
func FetchTokenList(baseURL string, whiteList []string) ([]Token, error) {
	url := fmt.Sprintf("%s/coins/list?include_platform=true", baseURL)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch token list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("CoinGecko API returned status: %s, body: %s", resp.Status, string(body))
	}

	var tokens []Token
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return nil, fmt.Errorf("failed to parse token list response: %w", err)
	}

	// Extract token IDs.
	var tokensFiltered []Token
	for _, token := range tokens {
		if stringInSlice(token.Symbol, whiteList) {
			tokensFiltered = append(tokensFiltered, token)
		}
	}

	return tokensFiltered, nil
}

func CalculateDailyPrices(baseURL, apiKey string, tokenIDs []Token) (map[string]float64, error) {
	// Batch tokens to limit the number of tokens per request.
	batches := batchTokens(tokenIDs, 50)

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        50,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	prices := make(map[string]float64)
	pricesFull := make(map[string]map[string]float64)
	var mu sync.Mutex

	bar := progressbar.Default(int64(len(batches)), "Processing token batches")

	var g errgroup.Group

	for _, batch := range batches {
		batch := batch

		g.Go(func() error {
			batchPricesFull, batchPrices, err := fetchBatchPricesWithRetry(client, baseURL, apiKey, batch)
			if err != nil {
				return fmt.Errorf("failed to fetch batch prices: %w", err)
			}

			// Merge batch prices into the result map.
			mu.Lock()
			for token, price := range batchPrices {
				prices[token] = price
			}
			mu.Unlock()

			mu.Lock()
			for symbol, prices := range batchPricesFull {
				if _, exists := pricesFull[symbol]; !exists {
					pricesFull[symbol] = make(map[string]float64)
				}

				for id, price := range prices {
					pricesFull[symbol][id] = price
				}
			}
			mu.Unlock()

			// Update progress bar.
			_ = bar.Add(1)

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Finish progress bar.
	_ = bar.Finish()

	prcs := make(map[string]float64)
	for k, v := range pricesFull {
		fmt.Println(k)
		fmt.Println("|")
		for kk, vv := range v {
			prcs[k] = vv
			fmt.Println("|---", kk, "=", vv)
			break
		}
	}

	return prcs, nil
}

func fetchBatchPricesWithRetry(client *http.Client, baseURL, apiKey string, tokens []Token) (map[string]map[string]float64, map[string]float64, error) {
	var tokenIDs []string
	for _, token := range tokens {
		tokenIDs = append(tokenIDs, token.ID)
	}
	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd", baseURL, strings.Join(tokenIDs, ","))

	var data map[string]map[string]float64
	for retries := 0; retries < 5; retries++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create request: %w", err)
		}

		if apiKey != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch token prices: %w", err)
		}
		defer resp.Body.Close()

		// Handle HTTP status codes.
		if resp.StatusCode == http.StatusOK {
			// Decode response if successful.
			err = json.NewDecoder(resp.Body).Decode(&data)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to decode response: %w", err)
			}
			break
		} else if resp.StatusCode == 429 {
			// Handle rate limit exceeded.
			retryAfter := parseRetryAfter(resp)
			if retries < 4 && retryAfter > 0 {
				time.Sleep(retryAfter)
				continue
			}
			return nil, nil, fmt.Errorf("CoinGecko API returned status: 429 Too Many Requests")
		} else {
			return nil, nil, fmt.Errorf("CoinGecko API returned status: %d", resp.StatusCode)
		}
	}

	// Extract prices.
	prices := make(map[string]float64)
	for token, value := range data {
		// fmt.Println("------------------------>", token, value)
		if usdPrice, ok := value["usd"]; ok {
			prices[token] = usdPrice
		}
	}

	pricesFull := buildPricesFull(tokens, prices)

	return pricesFull, prices, nil
}

func parseRetryAfter(resp *http.Response) time.Duration {
	retryAfterHeader := resp.Header.Get("Retry-After")
	if retryAfterHeader == "" {
		return 0
	}
	retryAfterSeconds, err := strconv.Atoi(retryAfterHeader)
	if err != nil {
		return 0
	}
	return time.Duration(retryAfterSeconds) * time.Second
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// batchTokens splits a list of tokens into smaller batches.
func batchTokens(tokens []Token, batchSize int) [][]Token {
	var batches [][]Token
	for batchSize < len(tokens) {
		tokens, batches = tokens[batchSize:], append(batches, tokens[:batchSize])
	}
	batches = append(batches, tokens)
	return batches
}

func buildPricesFull(tokens []Token, prices map[string]float64) map[string]map[string]float64 {
	pricesFull := make(map[string]map[string]float64)

	for _, token := range tokens {
		// Skip tokens with no price.
		price, found := prices[token.ID]
		if !found {
			continue
		}

		// Initialize the inner map for the symbol if it doesn't exist.
		if _, exists := pricesFull[token.Symbol]; !exists {
			pricesFull[token.Symbol] = make(map[string]float64)
		}

		// Add the price for the token ID under the token symbol.
		pricesFull[token.Symbol][token.ID] = price
	}

	return pricesFull
}
