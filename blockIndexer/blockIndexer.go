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
		logrus.WithFields(logrus.Fields{
			"scan_from": scanFrom,
			"scan_to":   scanTo,
			"scan_more": scanMore,
		}).WithError(err).Error("Fail call to resumeTask, scan stop")
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
				"scan_more": scanMore,
				"key":       proxy.LAST_BLOCK_INDEXED,
				"value":     scanTo}).WithError(err).Error("Fail call DbStoreSettingAsUint64, scan stop")
			return
		}

		if !scanMore {
			logrus.Info("ScanMore=false, scan stop")
			break
		}

		// update scanFrom , scanTo
		scanFrom, scanTo, err = nextTask(scanTo)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"scan_from": scanFrom,
				"scan_to":   scanTo,
				"scan_more": scanMore,
			}).WithError(err).Error("Fail call nextTask, scan stop")
			return
		}
	}
}

func nextTask(lastBlockIndexed uint64) (scanFrom uint64, scanTo uint64, err error) {
	scanFrom = lastBlockIndexed + 1
	scanTo = lastBlockIndexed + config.NUM_BLOCKS_SCAN_ONCE
	latestBlock, err := proxy.EthGetLatestBlockNumber()

	if err != nil {
		logrus.WithError(err).Error("Fail call EthGetLatestBlockNumber")
		return 0, 0, err
	}

	if scanFrom > latestBlock {
		return 0, 0, errors.New("Rich latest block and No block to scan")
	} else if scanTo > latestBlock {
		scanTo = latestBlock
	}

	return scanFrom, scanTo, nil
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

	// set default if userInput and db all 0
	if scanFrom == 0 && scanTo == 0 {
		scanFrom = 0
		scanTo = 20
	}

	// check scan task is all indexed
	// if no block indexed
	if dbLastIndexed > 0 {
		if scanTo <= dbLastIndexed {
			scanFrom, scanTo, err = nextTask(dbLastIndexed)
			return 0, 0, err
		} else {
			scanFrom = dbLastIndexed + 1
		}
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
			dbTranx, innerErr := proxy.EthFetchTranxByHash(v)
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
