package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/urfave/cli/v2"

	chaindataagg "github.com/atkachyshyn/chain-data-agg"
)

func main() {
	logger := chaindataagg.NewLogger("token-prices")

	app := &cli.App{
		Name:  "token-prices",
		Usage: "Fetch and calculate daily average token prices",
		Commands: []*cli.Command{
			{
				Name:   "run",
				Usage:  "Run the token price calculation",
				Action: runTokenPricesCommand(logger),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "output",
						Usage: "Path to save token price data",
					},
					&cli.StringSliceFlag{
						Name:     "tokens",
						Usage:    "Comma-separated list of tokens to process",
						Required: true,
					},
				},
			},
		},
	}

	// Run the app.
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runTokenPricesCommand(logger *slog.Logger) cli.ActionFunc {
	return func(c *cli.Context) error {
		// Load configuration.
		cfg, err := chaindataagg.LoadConfig()
		if err != nil {
			logger.Error("Failed to load configuration", slog.String("error", err.Error()))
			return err
		}

		bucketName := cfg.GCPBucketName

		//Ensure the token list is provided.
		if !c.IsSet("tokens") {
			logger.Error("Token list must be provided")
			return fmt.Errorf("token list is required")
		}
		tokens := c.StringSlice("tokens")
		logger.Info("Using provided token list", slog.Any("tokens", tokens))

		tokenList, err := chaindataagg.FetchTokenList(cfg.CoinGeckoBaseURL, tokens)
		if err != nil {
			return fmt.Errorf("token list is required")
		}
		fmt.Println(tokenList)

		// Ensure tokens are available.
		if len(tokens) == 0 {
			logger.Error("No tokens to process")
			return fmt.Errorf("no tokens available for processing")
		}

		// Calculate the date for the current run.
		today := time.Now().Format(chaindataagg.DateFormat)
		output := c.String("output")
		if output == "" {
			output = "prices/daily-token-prices-" + today + ".json"
		}

		logger.Info("Calculating prices for tokens", slog.Int("token_count", len(tokens)))

		// Fetch daily token prices.
		prices, err := chaindataagg.CalculateDailyPrices(cfg.CoinGeckoBaseURL, cfg.CoinGeckoAPIKey, tokenList)
		if err != nil {
			logger.Error("Failed to calculate token prices", slog.String("error", err.Error()))
			return err
		}

		// Serialize prices to JSON.
		serializedData, err := json.Marshal(prices)
		if err != nil {
			logger.Error("Failed to serialize token prices", slog.String("error", err.Error()))
			return err
		}

		// Upload the data to GCP.
		logger.Info("Uploading extracted data to GCP",
			slog.String("bucket", bucketName),
			slog.String("output", output),
		)
		if err := chaindataagg.UploadToBucket(bucketName, output, serializedData); err != nil {
			logger.Error("Failed to upload extracted data", slog.String("error", err.Error()))
			return err
		}

		logger.Info("Token prices saved successfully",
			slog.String("bucket", bucketName),
			slog.String("output", output),
		)
		return nil
	}
}
