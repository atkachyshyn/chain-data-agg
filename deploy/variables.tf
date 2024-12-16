variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "region" {
  description = "GCP Region"
  default     = "us-central1"
}

variable "bucket_name" {
  description = "Name of the GCP bucket"
  default     = "aggregator-data-bucket"
}

variable "clickhouse_host" {
  description = "ClickHouse Host URL"
  type        = string
}

variable "clickhouse_port" {
  description = "ClickHouse Port"
  default     = "9440"
}

variable "clickhouse_user" {
  description = "ClickHouse Username"
  type        = string
}

variable "clickhouse_password" {
  description = "ClickHouse Password"
  type        = string
}

variable "log_level" {
  description = "Log Level for the Application"
  default     = "INFO"
}

variable "sample_data_path" {
  description = "Path to the Sample Data File"
  default     = "sample_data/sample_data.csv"
}

variable "google_credentials" {
  description = "Path to Google Application Credentials"
  type        = string
}

variable "coingecko_api_key" {
  description = "API Key for CoinGecko"
  type        = string
}

variable "coingecko_base_url" {
  description = "Base URL for CoinGecko API"
  default     = "https://api.coingecko.com/api/v3"
}

variable "workers_num" {
  description = "Number of Workers for Parallel Processing"
  default     = 10
}

variable "tokens" {
  description = "List of tokens to fetch prices for"
  type        = string
  default     = "sfl,matic,usdc,usdc.e"
}

variable "token_prices_file" {
  description = "Output file to store token prices"
  type        = string
  default     = "token-prices.json"
}
