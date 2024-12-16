# Blockchain Data Aggregator for Marketplace Analytics

## Overview
The Blockchain Data Aggregator is a comprehensive tool for extracting, transforming, and loading (ETL) blockchain data for analytics purposes. It supports ingestion of transaction data, currency exchange rates, and other marketplace analytics, enabling data aggregation and loading into ClickHouse (extensible to other destinations) for further analysis.

This project provides two CLI applications:
- **Aggregator**: Handles the ETL pipeline (`extract`, `transform`, `load`) for data processing.
- **Token Prices**: Fetches and calculates token prices using external APIs.

---

## **Features**
- **Dynamic GCP Bucket Naming**: Bucket names follow the pattern `<bucket-name>-<project-id>` to ensure uniqueness across projects.
- **ETL Pipeline**: Supports full ETL processing (`extract`, `transform`, `load`) via a single command.
- **ClickHouse Schema Management**: Automates schema application to ClickHouse databases.
- **Test Data Upload**: Facilitates uploading test data to a GCP bucket.
- **Docker Integration**: Builds and pushes Docker images for deployment.
- **Terraform Integration**: Manages cloud infrastructure using Terraform.

---

## **Environment Variables**

The following environment variables must be defined in a `.env` file or exported into your environment:

### **GCP Configuration**
- `TF_VAR_project_id`: GCP project ID (e.g., `my-gcp-project`)
- `TF_VAR_region`: GCP region (e.g., `us-central1`)
- `TF_VAR_bucket_name`: Base bucket name (e.g., `aggregator-data`)
- `GOOGLE_APPLICATION_CREDENTIALS`: Path to your GCP service account JSON key (e.g., `path/to/service-account.json`)

### **ClickHouse Configuration**
- `CLICKHOUSE_HOST`: Hostname for ClickHouse (e.g., `clickhouse.example.com`)
- `CLICKHOUSE_PORT`: Port for ClickHouse (e.g., `9440`)
- `CLICKHOUSE_USER`: Username for ClickHouse (e.g., `default`)
- `CLICKHOUSE_PASSWORD`: Password for ClickHouse

### **CoinGecko Configuration**
- `COINGECKO_API_KEY`: API key for accessing CoinGecko APIs

### **Application Configuration**
- `LOG_LEVEL`: Log level for the application (e.g., `INFO`)

---

## **Setup**

### **Cloud Deployment Setup**

Before running the application locally or uploading test data, the required cloud infrastructure must be set up:

#### **Build and Push Docker Images**
Build and push Docker images for both services:
```bash
make build
make push
```

#### **Deploy Infrastructure**
Deploy the infrastructure using Terraform:
```bash
make deploy
```

---

### **Install Dependencies**
To install required dependencies:
```bash
make deps
```

### **Lint the Code**
To ensure code quality:
```bash
make lint
```

### **Build CLI Applications**
To build the `aggregator` and `token-prices` binaries locally:
```bash
make build-local
```

---

## **ClickHouse Schema Management**

Apply the schema to ClickHouse:
```bash
make apply-schema
```

Ensure the `schema/marketplace_analytics.sql` file contains the correct schema definition.

---

## **Upload Test Data**

Upload sample test data to the GCP bucket:
```bash
make upload-test-data
```
This will upload `sample_data/sample_data.csv` to the bucket `<bucket-name>-<project-id>`.

---

## **Running Locally**

### **Run Aggregator Commands**
Run specific aggregator commands with the appropriate parameters:
- **Extract**:
  ```bash
  make run-aggregator COMMAND="extract" INPUT="input.csv" OUTPUT="extracted.csv"
  ```
- **Transform**:
  ```bash
  make run-aggregator COMMAND="transform" INPUT="extracted.csv" OUTPUT="transformed.csv" RATES="rates.json"
  ```
- **Load**:
  ```bash
  make run-aggregator COMMAND="load" INPUT="transformed.csv" DESTINATION="ClickHouse"
  ```

### **Run Token Prices**
Fetch token prices:
```bash
make run-token-prices TOKENS="ETH,BTC" OUTPUT="prices.json"
```

### **Run the Full ETL Pipeline**
Run the full ETL pipeline sequentially:
```bash
make run-etl \
    EXTRACT_INPUT="input.csv" \
    EXTRACT_OUTPUT="extracted" \
    TRANSFORM_OUTPUT="transformed" \
    TRANSFORM_RATES="rates" \
    LOAD_DESTINATION="ClickHouse"
```

---

## **Cleanup**

Clean up the local environment:
```bash
make clean
```

---

## **Testing**

Run all tests:
```bash
make test
