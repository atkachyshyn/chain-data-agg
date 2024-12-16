package chaindataagg

import (
	"fmt"
	"strings"
)

type AggregatedData struct {
	Date           string
	ProjectID      string
	Transactions   int
	TotalVolumeUSD float64
}

func Transform(transactions []Transaction, currencyRates map[string]float64) ([]AggregatedData, error) {
	data := make(map[string]AggregatedData)

	for i, tx := range transactions {
		currencySymbol := strings.ToLower(tx.CurrencySymbol)
		date := strings.Split(tx.Timestamp, " ")[0]
		key := date + "_" + tx.ProjectID
		rate, ok := currencyRates[currencySymbol]
		if !ok {
			return nil, fmt.Errorf("missing exchange rate for %s", currencySymbol)
		}

		volumeUSD := tx.CurrencyValue * rate
		fmt.Println("[", i, "] ", tx.CurrencyValue, "*", rate, "=", volumeUSD)
		if _, exists := data[key]; !exists {
			data[key] = AggregatedData{
				Date:           date,
				ProjectID:      tx.ProjectID,
				Transactions:   0,
				TotalVolumeUSD: 0,
			}
		}

		entry := data[key]
		entry.Transactions++
		entry.TotalVolumeUSD += volumeUSD
		data[key] = entry
	}

	var aggregatedData []AggregatedData
	for _, entry := range data {
		aggregatedData = append(aggregatedData, entry)
	}

	return aggregatedData, nil
}
