package chaindataagg_test

import (
	"testing"

	chaindataagg "github.com/atkachyshyn/chain-data-agg"
)

func TestTransform(t *testing.T) {
	transactions := []chaindataagg.Transaction{
		{Timestamp: "2024-01-01T00:00:00Z", ProjectID: "1", CurrencySymbol: "BTC", CurrencyValue: 5000.0},
	}
	rates := map[string]float64{"btc": 1.0}

	t.Run("valid input", func(t *testing.T) {
		aggregated, err := chaindataagg.Transform(transactions, rates)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(aggregated) != 1 {
			t.Fatalf("expected 1 aggregated entry, got %d", len(aggregated))
		}
		if aggregated[0].TotalVolumeUSD != 5000.0 {
			t.Fatalf("expected total volume 5000.0, got %f", aggregated[0].TotalVolumeUSD)
		}
	})
}
