package chaindataagg

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	LogLevel           string
	ClickHouseHost     string
	ClickHousePort     string
	ClickHouseUser     string
	ClickHousePassword string
	SampleDataPath     string
	GCPBucketName      string
	GoogleCredentials  string
	CoinGeckoAPIKey    string
	CoinGeckoBaseURL   string
	WorkersNum         int
}

func LoadConfig() (*Config, error) {
	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
	}

	workersNum, err := strconv.Atoi(getEnv("WORKERS_NUM", "10"))
	if err != nil {
		return nil, fmt.Errorf("error extracting workers number: %w", err)
	}

	// Map required environment variables to configuration struct.
	return &Config{
		ClickHouseHost:     os.Getenv("CLICKHOUSE_HOST"),
		ClickHousePort:     os.Getenv("CLICKHOUSE_PORT"),
		ClickHouseUser:     os.Getenv("CLICKHOUSE_USER"),
		ClickHousePassword: os.Getenv("CLICKHOUSE_PASSWORD"),
		LogLevel:           getEnv("LOG_LEVEL", "INFO"),
		SampleDataPath:     getEnv("SAMPLE_DATA_PATH", "sample_data/sample_data.csv"),
		GCPBucketName:      os.Getenv("GCP_BUCKET_NAME"),
		GoogleCredentials:  os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
		CoinGeckoAPIKey:    os.Getenv("COINGECKO_API_KEY"),
		CoinGeckoBaseURL:   getEnv("COINGECKO_BASE_URL", "https://api.coingecko.com/api/v3"),
		WorkersNum:         workersNum,
	}, nil
}

// getEnv helper function to get an environment variable or use a fallback value.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func (c *Config) Validate() error {
	missing := []string{}

	// Check required fields.
	if c.ClickHouseHost == "" {
		missing = append(missing, "CLICKHOUSE_HOST")
	}
	if c.ClickHousePort == "" {
		missing = append(missing, "CLICKHOUSE_PORT")
	}
	if c.ClickHouseUser == "" {
		missing = append(missing, "CLICKHOUSE_USER")
	}
	if c.ClickHousePassword == "" {
		missing = append(missing, "CLICKHOUSE_PASSWORD")
	}
	if c.SampleDataPath == "" {
		missing = append(missing, "SAMPLE_DATA_PATH")
	}
	if c.GCPBucketName == "" {
		missing = append(missing, "GCP_BUCKET_NAME")
	}
	if c.GoogleCredentials == "" {
		missing = append(missing, "GOOGLE_APPLICATION_CREDENTIALS")
	}
	if c.CoinGeckoAPIKey == "" {
		missing = append(missing, "COINGECKO_API_KEY")
	}
	if c.CoinGeckoBaseURL == "" {
		missing = append(missing, "COINGECKO_BASE_URL")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missing)
	}

	return nil
}

func (c *Config) PrintExportedVars() {
	fmt.Println("Exported Terraform Variables:")
	fmt.Printf("TF_VAR_project_id=%s\n", c.GCPBucketName)
	fmt.Printf("TF_VAR_clickhouse_host=%s\n", c.ClickHouseHost)
	fmt.Printf("TF_VAR_clickhouse_port=%s\n", c.ClickHousePort)
	fmt.Printf("TF_VAR_clickhouse_user=%s\n", c.ClickHouseUser)
	fmt.Printf("TF_VAR_clickhouse_password=%s\n", c.ClickHousePassword)
	fmt.Printf("TF_VAR_google_credentials=%s\n", c.GoogleCredentials)
	fmt.Printf("TF_VAR_coingecko_api_key=%s\n", c.CoinGeckoAPIKey)
}
