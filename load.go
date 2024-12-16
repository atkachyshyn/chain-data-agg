package chaindataagg

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

func Load(data []AggregatedData, host, password string) error {
	db, err := connect(host, password)
	if err != nil {
		return err
	}
	defer db.Close()

	query := `
		INSERT INTO marketplace_analytics (date, project_id, transactions, total_volume_usd)
		VALUES (?, ?, ?, ?)
	`

	ctx := context.Background()
	for _, entry := range data {
		// _, err := db.Exec(query, entry.Date, entry.ProjectID, entry.Transactions, entry.TotalVolumeUSD)
		_, err := db.Query(ctx, query, entry.Date, entry.ProjectID, entry.Transactions, entry.TotalVolumeUSD)
		if err != nil {
			return fmt.Errorf("failed to insert data: %w", err)
		}
	}

	return nil
}

func connect(host, password string) (driver.Conn, error) {
	var (
		ctx       = context.Background()
		conn, err = clickhouse.Open(&clickhouse.Options{
			Addr: []string{fmt.Sprintf("%s:9440", host)},
			Auth: clickhouse.Auth{
				Database: "analytics",
				Username: "default",
				Password: password,
			},
			ClientInfo: clickhouse.ClientInfo{
				Products: []struct {
					Name    string
					Version string
				}{
					{Name: "clickhouse-go-client", Version: "0.1"},
				},
			},

			Debugf: func(format string, v ...interface{}) {
				fmt.Printf(format, v)
			},
			TLS: &tls.Config{
				InsecureSkipVerify: true,
			},
		})
	)

	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return nil, err
	}
	return conn, nil
}
