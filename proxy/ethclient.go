package proxy

import (
	"context"
	"math/big"
	"myModule/config"
	"myModule/model"
	"myModule/utils"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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
	result.Blocks = make([]*model.DbBlockBase, numBlocks)

	_ = utils.BatchUint64(numBlocks, batchSize, func(start, end uint64) error {
		wg.Add(1)
		go fetchHeaderRangeByNumber(start+blockStartIndex, end+blockStartIndex, result.Blocks[start:end+1], &wg)
		return nil
	})

	wg.Wait()

	return result, nil
}

func EthFetchBlockByNumber(blockNum uint64) (*model.BlockTranx, error) {
	n := new(big.Int).SetUint64(blockNum)
	block, err := client.BlockByNumber(ctx, n)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"block_num": n,
		}).WithError(err).Error("Failed fetching block")
		return nil, err
	}

	result := &model.BlockTranx{
		DbBlockBase: model.DbBlockBase{
			Num:        block.NumberU64(),
			Hash:       block.Hash().String(),
			Time:       block.Time(),
			ParentHash: block.ParentHash().String(),
		},
		TranxHash: make([]string, block.Transactions().Len()),
	}
	for i, v := range block.Transactions() {
		result.TranxHash[i] = v.Hash().String()
	}

	return result, nil
}

func fetchHeaderRangeByNumber(blockStart uint64, blockEnd uint64, inOut []*model.DbBlockBase, wg *sync.WaitGroup) {
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
				"block_num": n,
			}).WithError(err).Error("Failed fetching header")
			// continue
		} else {
			inOut[i-blockStart] = &model.DbBlockBase{
				Num:        header.Number.Uint64(),
				Hash:       header.Hash().String(),
				Time:       header.Time,
				ParentHash: header.ParentHash.String(),
			}
		}
	}
}

/**
 * @return DbTranx The filed BlockNum, BlockHash will not set.
 */
func EthFetchTranxByBash(hash string) (*model.DbTranx, error) {
	if strings.HasPrefix(hash, "0x") {
		hash = hash[2:]
	}

	// fetch tranx by hash
	tx, _, err := client.TransactionByHash(ctx, common.HexToHash(hash))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tranx_hash": hash,
		}).WithError(err).Error("Failed fetching tranx")
		return nil, err
	}

	// get chain id for EIP155 Signer
	chainID, err := client.NetworkID(ctx)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tranx_hash": hash,
		}).WithError(err).Error("Failed fetching chainID")
		return nil, err
	}

	// tranx to msg to get "from" field
	msg, err := tx.AsMessage(types.NewEIP155Signer(chainID), nil)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tranx_hash": hash,
			"chain_id":   chainID,
		}).WithError(err).Error("Failed call AsMessage")
		return nil, err
	}

	// fetch receipt to get log
	receipt, err := client.TransactionReceipt(ctx, tx.Hash())
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tranx_hash": hash,
			"chain_id":   chainID,
		}).WithError(err).Error("Failed call TransactionReceipt")
		return nil, err
	}

	// logrus.Infof("hash-string", msg.From().Hash().String()) => ok
	// logrus.Infof("hash-hex", msg.From().Hash().Hex()) => ok
	// logrus.Infof("hex", msg.From().Hex()) => not right value
	// logrus.Infof("string", msg.From().String()) => not right value

	result := &model.DbTranx{
		Hash:  tx.Hash().String(),
		From:  msg.From().Hash().String(),
		To:    tx.To().Hash().String(),
		Nonce: tx.Nonce(),
		Data:  common.BytesToHash(tx.Data()).String(),
		Value: tx.Value().String(),
		Logs:  make([]model.DbTranxLog, len(receipt.Logs)),
	}

	for i, v := range receipt.Logs {
		result.Logs[i] = model.DbTranxLog{
			TranxHash: result.Hash,
			Index:     v.Index,
			Data:      common.BytesToHash(v.Data).String(),
		}
	}

	return result, nil
}
