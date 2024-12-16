package chaindataagg

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type Transaction struct {
	Timestamp      string
	Event          string
	ProjectID      string
	CurrencySymbol string
	CurrencyValue  float64
}

func Extract(inputData []byte, workerCount int) ([]Transaction, error) {
	reader := csv.NewReader(bytes.NewReader(inputData))

	// Read header row to skip it.
	_, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Read all records (small performance trade-off to distribute work).
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read records: %w", err)
	}

	// Channel for records and results.
	recordChan := make(chan []string, len(records))
	resultChan := make(chan Transaction, len(records))
	errorChan := make(chan error, len(records))
	wg := &sync.WaitGroup{}

	// Start worker goroutines.
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go processRecords(recordChan, resultChan, errorChan, wg)
	}

	// Send records to workers.
	go func() {
		for _, record := range records {
			recordChan <- record
		}
		close(recordChan)
	}()

	// Wait for workers to finish.
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// Collect results.
	var transactions []Transaction
	for transaction := range resultChan {
		transactions = append(transactions, transaction)
	}

	// Collect errors.
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}
	if len(errors) > 0 {
		return nil, fmt.Errorf("errors occurred during extraction: %v", errors)
	}

	return transactions, nil
}

func processRecords(recordChan <-chan []string, resultChan chan<- Transaction, errorChan chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	for record := range recordChan {
		transaction, err := processRecord(record)
		if err != nil {
			errorChan <- err
			continue
		}
		resultChan <- transaction
	}
}

func processRecord(record []string) (Transaction, error) {
	props := record[14]
	nums := record[15]

	currencySymbol := extractJSONValue(props, `"currencySymbol"`)

	currencyValueDecimalStr := extractJSONValue(nums, `"currencyValueDecimal"`)
	currencyValueDecimal, err := strconv.ParseFloat(currencyValueDecimalStr, 64)
	if err != nil {
		return Transaction{}, fmt.Errorf("failed to parse currency value: %w", err)
	}

	// Create the transaction.
	return Transaction{
		Timestamp:      record[1], // ts
		Event:          record[2], // event
		ProjectID:      record[3], // project_id
		CurrencySymbol: currencySymbol,
		CurrencyValue:  currencyValueDecimal,
	}, nil
}

func extractJSONValue(jsonStr, key string) string {
	start := strings.Index(jsonStr, key)
	if start == -1 {
		return ""
	}

	start += len(key) + 2
	if jsonStr[start] == '"' {
		start++
	}

	end := strings.Index(jsonStr[start:], `"`)
	if end == -1 {
		end = len(jsonStr[start:])
	}
	return jsonStr[start : start+end]
}
