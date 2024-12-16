package chaindataagg

import (
	"encoding/csv"
	"log/slog"
	"os"
	"strconv"
)

const (
	DateFormat = "YYYY-MM-DD"
)

// NewLogger constructs new logger.
func NewLogger(stage string) *slog.Logger {
	leveler := new(slog.LevelVar)
	leveler.Set(slog.LevelInfo)
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: leveler,
	})
	baseLogger := slog.New(handler)
	return baseLogger.With(slog.String("stage", stage))
}

// ParseLevel parses log level from string value.
func ParseLevel(s string) (slog.Level, error) {
	var level slog.Level
	var err = level.UnmarshalText([]byte(s))
	return level, err
}

func ParseCSV(filePath string) ([]Transaction, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var transactions []Transaction
	for _, record := range records[1:] {
		value, _ := strconv.ParseFloat(record[4], 64)
		transactions = append(transactions, Transaction{
			Timestamp:      record[0],
			Event:          record[1],
			ProjectID:      record[2],
			CurrencySymbol: record[3],
			CurrencyValue:  value,
		})
	}

	return transactions, nil
}
