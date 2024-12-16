package chaindataagg_test

import (
	"os"
	"testing"

	chaindataagg "github.com/atkachyshyn/chain-data-agg"
)

func TestParseCSV(t *testing.T) {
	data := `timestamp,event,project_id,currencySymbol,currencyValue
2024-01-01T00:00:00Z,event1,1,BTC,5000.0`
	filePath := "test.csv"
	err := os.WriteFile(filePath, []byte(data), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	defer os.Remove(filePath)

	t.Run("valid parse", func(t *testing.T) {
		transactions, err := chaindataagg.ParseCSV(filePath)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(transactions) != 1 {
			t.Fatalf("expected 1 transaction, got %d", len(transactions))
		}
		if transactions[0].CurrencySymbol != "BTC" {
			t.Fatalf("expected currency symbol BTC, got %s", transactions[0].CurrencySymbol)
		}
	})
}
