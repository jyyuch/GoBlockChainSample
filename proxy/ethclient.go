package proxy

import (
	"context"
	"math/big"
	"myModule/config"
	"myModule/model"
	"myModule/utils"
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
)

var client *ethclient.Client
var ctx = context.Background()

func init() {
	tmpClient, err := ethclient.Dial(config.RPC_ENDPOINT)
	if err != nil {
		panic(err)
	}

	client = tmpClient
}

func EthGetLatestBlocks(numLatestBlocks uint64) (*model.ResponseBlocks, error) {
	latestBlockNum, err := client.BlockNumber(context.Background())
	if err != nil {
		return nil, err
	}

	// batch
	// ex: numLatestBlocks = 20, latestBlock = #120, want get block #100~#120
	blockStartIndex := latestBlockNum - numLatestBlocks
	numBlocks := numLatestBlocks
	if latestBlockNum < numLatestBlocks {
		blockStartIndex = 0
		numBlocks = latestBlockNum + 1
	}

	var batchSize uint64
	if numBlocks > config.NUM_ROUTINE_TO_SCAN {
		batchSize = numBlocks / config.NUM_ROUTINE_TO_SCAN
	} else {
		batchSize = numBlocks
	}
	var wg sync.WaitGroup

	result := &model.ResponseBlocks{}
	result.Blocks = make([]*model.BlockBase, numBlocks)

	_ = utils.BatchUint64(numBlocks, batchSize, func(start, end uint64) error {
		wg.Add(1)
		go fetchBlockByNumber(start+blockStartIndex, end+blockStartIndex, result.Blocks[start:end+1], &wg)
		return nil
	})

	wg.Wait()

	return result, nil
}

func fetchBlockByNumber(blockStart uint64, blockEnd uint64, inOut []*model.BlockBase, wg *sync.WaitGroup) {
	defer func() {
		if wg != nil {
			wg.Done()
		}
	}()

	for i := blockStart; i <= blockEnd; i++ {
		n := new(big.Int).SetUint64(i)
		header, err := client.HeaderByNumber(ctx, n)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"block_num": header.Number.Uint64(),
			}).WithError(err).Error("Failed fetching block")
		} else {
			inOut[i-blockStart] = &model.BlockBase{
				Num:        header.Number.Uint64(),
				Hash:       header.Hash().String(),
				Time:       header.Time,
				ParentHash: header.ParentHash.String(),
			}
		}
	}
}
