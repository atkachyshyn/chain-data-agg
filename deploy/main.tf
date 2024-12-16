provider "google" {
  credentials = file("${path.module}/service-account.json")
  project     = var.project_id
  region      = var.region
}

# GCP Bucket for Intermediate Data.
resource "google_storage_bucket" "aggregator_bucket" {
  name     = "${var.bucket_name}-${var.project_id}"
  location = var.region
}

# Google Cloud Run Service for Token Prices CLI.
resource "google_cloud_run_service" "token_prices_service" {
  name     = "token-prices"
  location = var.region

  template {
    spec {
      containers {
        image = "gcr.io/${var.project_id}/token-prices:latest"
        command = ["/bin/sh", "-c"]
        args    = ["./token-prices run --tokens=$TOKENS --output=$TOKEN_PRICES_FILE"]
        env {
          name  = "COINGECKO_API_KEY"
          value = var.coingecko_api_key
        }
        env {
          name  = "COINGECKO_BASE_URL"
          value = var.coingecko_base_url
        }
        env {
          name  = "TOKENS"
          value = var.tokens
        }
        env {
          name  = "TOKEN_PRICES_FILE"
          value = var.token_prices_file
        }
      }
    }
  }

  metadata {
    annotations = {
      "autoscaling.knative.dev/maxScale" = "3"
    }
  }
}

resource "google_cloud_scheduler_job" "token_prices_job" {
  name        = "fetch-token-prices"
  schedule    = "0 0 * * *"
  time_zone   = "UTC"

  http_target {
    http_method = "POST"
    uri         = "${google_cloud_run_service.token_prices_service.status[0].url}"
    oidc_token {
      service_account_email = google_service_account.scheduler_sa.email
    }
  }
}

# Service Account for Cloud Scheduler.
resource "google_service_account" "scheduler_sa" {
  account_id   = "token-prices-scheduler-sa"
  display_name = "Service Account for Token Prices Scheduler"
}

# IAM Role Assignment for Scheduler.
resource "google_project_iam_member" "scheduler_sa_role" {
  project = var.project_id
  role    = "roles/cloudscheduler.serviceAgent"
  member  = "serviceAccount:${google_service_account.scheduler_sa.email}"
}

# Google Cloud Run Service for Aggregator CLI
resource "google_cloud_run_service" "aggregator_service" {
  name     = "aggregator-service"
  location = var.region

  template {
    spec {
      containers {
        image = "gcr.io/${var.project_id}/aggregator:latest"
        command = ["./aggregator"]
        env {
          name  = "CLICKHOUSE_HOST"
          value = var.clickhouse_host
        }
        env {
          name  = "CLICKHOUSE_PORT"
          value = var.clickhouse_port
        }
        env {
          name  = "CLICKHOUSE_USER"
          value = var.clickhouse_user
        }
        env {
          name  = "CLICKHOUSE_PASSWORD"
          value = var.clickhouse_password
        }
        env {
          name  = "LOG_LEVEL"
          value = var.log_level
        }
        env {
          name  = "SAMPLE_DATA_PATH"
          value = var.sample_data_path
        }
        env {
          name  = "GCP_BUCKET_NAME"
          value = var.bucket_name
        }
        env {
          name  = "GOOGLE_APPLICATION_CREDENTIALS"
          value = var.google_credentials
        }
        env {
          name  = "COINGECKO_API_KEY"
          value = var.coingecko_api_key
        }
        env {
          name  = "COINGECKO_BASE_URL"
          value = var.coingecko_base_url
        }
        env {
          name  = "WORKERS_NUM"
          value = var.workers_num
        }
      }
    }
  }

  metadata {
    annotations = {
      "autoscaling.knative.dev/maxScale" = "3"
    }
  }
}

resource "google_project_iam_member" "cloud_run_invoker" {
  project = var.project_id
  role    = "roles/run.invoker"
  member  = "allUsers"
}
