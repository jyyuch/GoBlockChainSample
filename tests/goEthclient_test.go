package tests

import (
	"context"
	"fmt"
	"myModule/config"
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
)

func TestGoEthclient(t *testing.T) {
	assert := assert.New(t)

	client, err := ethclient.Dial(config.RPC_ENDPOINT)
	assert.Nil(err)

	latestBlock, err := client.BlockNumber(context.Background())
	assert.Nil(err)

	fmt.Printf("latestBlock=%d", latestBlock)
}
