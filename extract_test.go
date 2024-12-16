package chaindataagg_test

import (
	"testing"

	chaindataagg "github.com/atkachyshyn/chain-data-agg"
)

func TestExtract(t *testing.T) {
	sampleData := `"app","ts","event","project_id","source","ident","user_id","session_id","country","device_type","device_os","device_os_ver","device_browser","device_browser_ver","props","nums"
"seq-market","2024-04-15 02:15:07.167","BUY_ITEMS","4974","","1","0896ae95dcaeee38e83fa5c43bef99780d7b2be23bcab36214","5d8afd8fec2fbf3e","DE","desktop","linux","x86_64","chrome","122.0.0.0","{""tokenId"":""215"",""txnHash"":""0xd919290e80df271e77d1cbca61f350d2727531e0334266671ec20d626b2104a2"",""chainId"":""137"",""collectionAddress"":""0x22d5f9b75c524fec1d6619787e582644cd4d7422"",""currencyAddress"":""0xd1f9c58e33933a993a3891f8acfe05a68e1afc05"",""currencySymbol"":""SFL"",""marketplaceType"":""amm"",""requestId"":""""}","{""currencyValueDecimal"":""0.6136203411678249"",""currencyValueRaw"":""613620341167824900""}"`

	// Convert the sample data into a byte slice
	inputBytes := []byte(sampleData)

	t.Run("valid input", func(t *testing.T) {
		// Call the Extract function
		transactions, err := chaindataagg.Extract(inputBytes, 1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Validate the number of extracted transactions
		if len(transactions) != 1 {
			t.Fatalf("expected 1 transaction, got %d", len(transactions))
		}

		// Validate content of the extracted transaction
		transaction := transactions[0]
		if transaction.Event != "BUY_ITEMS" {
			t.Fatalf("expected event BUY_ITEMS, got %s", transaction.Event)
		}
		if transaction.ProjectID != "4974" {
			t.Fatalf("expected project_id 4974, got %s", transaction.ProjectID)
		}
		if transaction.CurrencySymbol != "SFL" {
			t.Fatalf("expected currency symbol SFL, got %s", transaction.CurrencySymbol)
		}
	})
}
