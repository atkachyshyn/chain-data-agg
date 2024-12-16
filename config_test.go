package chaindataagg_test

import (
	"os"
	"testing"

	chaindataagg "github.com/atkachyshyn/chain-data-agg"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	os.Setenv("CLICKHOUSE_HOST", "localhost")
	os.Setenv("CLICKHOUSE_PORT", "9440")
	os.Setenv("WORKERS_NUM", "5")
	config, err := chaindataagg.LoadConfig()
	require.NoError(t, err)
	require.Equal(t, "localhost", config.ClickHouseHost)
	require.Equal(t, "9440", config.ClickHousePort)
	require.Equal(t, 5, config.WorkersNum)
}

func TestValidateConfig(t *testing.T) {
	config := chaindataagg.Config{
		ClickHouseHost: "",
		ClickHousePort: "9440",
		WorkersNum:     5,
	}
	err := config.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required environment variables")
}
