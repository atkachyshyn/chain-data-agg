CREATE TABLE IF NOT EXISTS marketplace_analytics (
    date Date,
    project_id String,
    transactions UInt32,
    total_volume_usd Float64
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (date, project_id);
