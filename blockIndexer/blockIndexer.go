package blockIndexer

import (
	"myModule/config"
	"myModule/model"
	"myModule/proxy"
	"myModule/utils"

	"errors"
	"sync"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	jobIsRunning   bool
	JobIsrunningMu sync.Mutex
)

func StartScanRoutine(blockFrom uint64, blockTo uint64, scanMore bool) (start bool) {
	JobIsrunningMu.Lock()
	start = !jobIsRunning
	jobIsRunning = true
	JobIsrunningMu.Unlock()
	if start {
		go scanRoutine(blockFrom, blockTo, scanMore)
	}

	return start
}

func scanRoutine(blockFrom uint64, blockTo uint64, scanMore bool) {
	defer func() {
		JobIsrunningMu.Lock()
		jobIsRunning = false
		JobIsrunningMu.Unlock()
	}()

	scanFrom, scanTo, err := resumeTask(blockFrom, blockTo)
	if err != nil {
		logrus.Error(err.Error())
		return
	}

	for {
		storeTask(scanFrom, scanTo)

		numBlocks := scanTo - scanFrom + 1
		var batchSize uint64
		if numBlocks > config.NUM_ROUTINE_TO_SCAN {
			batchSize = numBlocks / config.NUM_ROUTINE_TO_SCAN
		} else {
			batchSize = numBlocks
		}

		var wg sync.WaitGroup

		_ = utils.BatchUint64(numBlocks, batchSize, func(start, end uint64) error {
			wg.Add(1)
			go scanBlockRangeByNumber(start+scanFrom, end+scanFrom, &wg)
			return nil
		})

		wg.Wait()

		err = proxy.DbStoreSettingAsUint64(proxy.LAST_BLOCK_INDEXED, scanTo)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"scan_from": scanFrom,
				"scan_to":   scanTo,
				"key":       proxy.LAST_BLOCK_INDEXED,
				"value":     scanTo}).WithError(err).Error("Fail call DbStoreSettingAsUint64")
			return
		}

		if !scanMore {
			break
		}

		// update scanFrom , scanTo
		scanFrom, scanTo = nextTask(scanTo)
	}
}

func nextTask(lastBlockIndexed uint64) (scanFrom uint64, scanTo uint64) {
	scanFrom = lastBlockIndexed + 1
	scanTo = lastBlockIndexed + config.NUM_BLOCKS_SCAN_ONCE
	return scanFrom, scanTo
}

func resumeTask(userScanFrom uint64, userScanTo uint64) (scanFrom uint64, scanTo uint64, err error) {
	// load db already index, scan_from, scan_to
	dbScanFrom, err := proxy.DbLoadSettingAsUint64(proxy.TASK_SCAN_FROM, 0)
	if err != nil {
		return 0, 0, err
	}

	dbScanTo, err := proxy.DbLoadSettingAsUint64(proxy.TASK_SCAN_TO, 0)
	if err != nil {
		return 0, 0, err
	}

	dbLastIndexed, err := proxy.DbLoadSettingAsUint64(proxy.LAST_BLOCK_INDEXED, 0)
	if err != nil {
		return 0, 0, err
	}

	scanFrom = dbScanFrom
	scanTo = dbScanTo
	// if db no pre task => use userInput
	if dbScanFrom == 0 && dbScanTo == 0 {
		scanFrom = userScanFrom
		scanTo = userScanTo
	}

	// check scan task is all indexed
	if scanTo <= dbLastIndexed {
		scanFrom, scanTo = nextTask(dbLastIndexed)
	} else {
		scanFrom = dbLastIndexed + 1
	}

	return scanFrom, scanTo, nil
}

func storeTask(blockFrom uint64, blockTo uint64) {
	proxy.DbStoreSettingAsUint64(proxy.TASK_SCAN_FROM, blockFrom)
	proxy.DbStoreSettingAsUint64(proxy.TASK_SCAN_TO, blockTo)
}

func scanBlockRangeByNumber(blockStart uint64, blockEnd uint64, wg *sync.WaitGroup) {
	defer func() {
		if wg != nil {
			wg.Done()
		}
	}()

	for scanBlock := blockStart; scanBlock <= blockEnd; scanBlock++ {
		// check if new block to scan
		err := proxy.DbLoad(&model.DbBlock{}, scanBlock)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.WithFields(logrus.Fields{"block_num": scanBlock}).WithError(err).Error("DbLoad BlockTranx error")
			continue
		} else if err == nil {
			// db already exist
			continue
		}

		blockTranx, err := proxy.EthFetchBlockByNumber(scanBlock)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"block_num": scanBlock,
			}).WithError(err).Error("Failed call EthFetchBlockByNumber")
			continue
		}

		dbBlock := &model.DbBlock{
			DbBlockBase:  blockTranx.DbBlockBase,
			Transactions: make([]model.DbTranx, len(blockTranx.TranxHash)),
		}

		for j, v := range blockTranx.TranxHash {
			dbTranx, innerErr := proxy.EthFetchTranxByBash(v)
			if innerErr != nil {
				logrus.WithFields(logrus.Fields{
					"block_num":  scanBlock,
					"tranx_hash": v,
				}).WithError(err).Error("Failed call EthFetchTranxByBash")
				err = innerErr
				break
			}

			// fill BlockHash, BlockNum
			dbTranx.BlockHash = dbBlock.Hash
			dbTranx.BlockNum = dbBlock.Num
			dbBlock.Transactions[j] = *dbTranx
		}

		if err != nil {
			continue
		}

		// store back
		err = proxy.DbStore(dbBlock)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"block_num": scanBlock,
			}).WithError(err).Error("Failed DbStore")
			// continue
		}
	}
}
