package config

/**
 * config.go could use viper to get var from env for k8s deployment.
 */
const (
	RPC_ENDPOINT         = "https://data-seed-prebsc-2-s3.binance.org:8545/"
	NUM_ROUTINE_TO_SCAN  = 10
	NUM_BLOCKS_SCAN_ONCE = 20
	PG_DSN               = "host=localhost user=postgres password=mysecretpassword dbname=postgres port=5432"
)
