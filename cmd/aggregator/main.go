package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"os"

	chaindataagg "github.com/atkachyshyn/chain-data-agg"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "aggregator",
		Usage: "Blockchain Data Aggregator",
		Commands: []*cli.Command{
			{
				Name:    "extract",
				Aliases: []string{"e"},
				Usage:   "Extract data from the source",
				Action:  extractAction(chaindataagg.NewLogger("extract")),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "input",
						Aliases:  []string{"i"},
						Usage:    "Path to input file (e.g., CSV)",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "output",
						Aliases:  []string{"o"},
						Usage:    "Path to save extracted data",
						Required: true,
					},
				},
			},
			{
				Name:    "transform",
				Aliases: []string{"t"},
				Usage:   "Transform extracted data",
				Action:  transformAction(chaindataagg.NewLogger("transform")),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "input",
						Aliases:  []string{"i"},
						Usage:    "Path to extracted data",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "output",
						Aliases:  []string{"o"},
						Usage:    "Path to save transformed data",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "rates",
						Aliases:  []string{"r"},
						Usage:    "Path to currency rates",
						Required: true,
					},
				},
			},
			{
				Name:    "load",
				Aliases: []string{"l"},
				Usage:   "Load transformed data into the destination",
				Action:  loadAction(chaindataagg.NewLogger("load")),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "input",
						Aliases:  []string{"i"},
						Usage:    "Path to transformed data",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "destination",
						Aliases: []string{"d"},
						Usage:   "Target destination (e.g., ClickHouse, BigQuery)",
						Value:   "ClickHouse",
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func extractAction(logger *slog.Logger) cli.ActionFunc {
	return func(c *cli.Context) error {
		// Load configuration.
		cfg, err := chaindataagg.LoadConfig()
		if err != nil {
			return err
		}

		// Input and output parameters.
		input := c.String("input")
		output := c.String("output")
		bucketName := cfg.GCPBucketName
		logger.Info("Starting data extraction",
			slog.String("input", input),
			slog.String("output", output),
		)

		// Step 1: Download file from GCP bucket.
		logger.Info("Downloading input file from GCP", slog.String("bucket", bucketName), slog.String("input", input))
		data, err := chaindataagg.DownloadFromBucket(bucketName, input)
		if err != nil {
			logger.Error("Failed to download input file from GCP", slog.String("error", err.Error()))
			return err
		}

		// Step 2: Extract transactions from the downloaded data.
		logger.Info("Extracting data")
		transactions, err := chaindataagg.Extract(data, cfg.WorkersNum)
		if err != nil {
			logger.Error("Failed to extract data", slog.String("error", err.Error()))
			return err
		}

		// Step 3: Serialize and upload the results to GCP.
		logger.Info("Serializing extracted data")
		serializedData, err := json.Marshal(transactions)
		if err != nil {
			logger.Error("Failed to serialize extracted data", slog.String("error", err.Error()))
			return err
		}

		logger.Info("Uploading extracted data to GCP", slog.String("bucket", bucketName), slog.String("output", output))
		if err := chaindataagg.UploadToBucket(bucketName, output, serializedData); err != nil {
			logger.Error("Failed to upload extracted data", slog.String("error", err.Error()))
			return err
		}

		logger.Info("Data extraction completed successfully",
			slog.String("output", output),
		)
		return nil
	}
}

func transformAction(logger *slog.Logger) cli.ActionFunc {
	return func(c *cli.Context) error {
		// Load configuration.
		cfg, err := chaindataagg.LoadConfig()
		if err != nil {
			return err
		}
		bucketName := cfg.GCPBucketName
		input := c.String("input")
		output := c.String("output")
		ratesPath := c.String("rates")
		logger.Info("Starting data transformation", slog.String("input", input), slog.String("output", output), slog.String("ratesPath", ratesPath))

		// Download extracted data from GCP.
		data, err := chaindataagg.DownloadFromBucket(bucketName, input)
		if err != nil {
			logger.Error("Failed to download extracted data", slog.String("error", err.Error()))
			return err
		}
		rates, err := chaindataagg.DownloadFromBucket(bucketName, ratesPath)
		if err != nil {
			logger.Error("Failed to download rates", slog.String("error", err.Error()))
			return err
		}

		var transactions []chaindataagg.Transaction
		if err := json.Unmarshal(data, &transactions); err != nil {
			logger.Error("Failed to deserialize extracted data", slog.String("error", err.Error()))
			return err
		}

		var currencyRates map[string]float64
		err = json.Unmarshal(rates, &currencyRates)
		if err != nil {
			return err
		}
		logger.Info("Currency rates", slog.Any("ratesPath", currencyRates))

		// Transform data.
		aggregatedData, err := chaindataagg.Transform(transactions, currencyRates)
		if err != nil {
			logger.Error("Failed to transform data", slog.String("error", err.Error()))
			return err
		}

		// Serialize and upload transformed data to GCP.
		data, err = json.Marshal(aggregatedData)
		if err != nil {
			logger.Error("Failed to serialize transformed data", slog.String("error", err.Error()))
			return err
		}

		if err := chaindataagg.UploadToBucket(bucketName, output, data); err != nil {
			logger.Error("Failed to upload transformed data", slog.String("error", err.Error()))
			return err
		}

		logger.Info("Data transformation completed", slog.String("output", output))
		return nil
	}
}

func loadAction(logger *slog.Logger) cli.ActionFunc {
	return func(c *cli.Context) error {
		// Load configuration.
		cfg, err := chaindataagg.LoadConfig()
		if err != nil {
			return err
		}
		bucketName := cfg.GCPBucketName
		input := c.String("input")
		destination := c.String("destination")
		logger.Info("Starting data load", slog.String("input", input), slog.String("destination", destination))

		// Download transformed data from GCP.
		data, err := chaindataagg.DownloadFromBucket(bucketName, input)
		if err != nil {
			logger.Error("Failed to download transformed data", slog.String("error", err.Error()))
			return err
		}

		var aggregatedData []chaindataagg.AggregatedData
		if err := json.Unmarshal(data, &aggregatedData); err != nil {
			logger.Error("Failed to deserialize transformed data", slog.String("error", err.Error()))
			return err
		}

		// Insert data into ClickHouse.
		if destination == "ClickHouse" {
			err = chaindataagg.Load(aggregatedData, cfg.ClickHouseHost, cfg.ClickHousePassword)
			if err != nil {
				logger.Error("Failed to insert data into ClickHouse", slog.String("error", err.Error()))
				return err
			}
		}

		logger.Info("Data load completed", slog.String("destination", destination))
		return nil
	}
}
