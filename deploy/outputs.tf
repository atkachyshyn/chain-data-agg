output "bucket_name" {
  description = "GCP Bucket Name"
  value       = google_storage_bucket.aggregator_bucket.name
}

output "token_prices_service_url" {
  description = "URL for the Token Prices Service"
  value       = google_cloud_run_service.token_prices_service.status[0].url
}

output "aggregator_service_url" {
  description = "URL for the Aggregator Service"
  value       = google_cloud_run_service.aggregator_service.status[0].url
}
