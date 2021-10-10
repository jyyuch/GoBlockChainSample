package proxy

import (
	"myModule/config"
	"myModule/model"

	"errors"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

const (
	LAST_BLOCK_INDEXED = "last_block_indexed"
	TASK_SCAN_FROM     = "task_scan_from"
	TASK_SCAN_TO       = "task_scan_to"
)

var db *gorm.DB

func init() {
	var err error
	db, err = gorm.Open(postgres.New(postgres.Config{
		DSN: config.PG_DSN,
	}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"pg_dsn": config.PG_DSN,
		}).WithError(err).Fatal("failed to connect database")
	}

	err = db.AutoMigrate(model.TablesMigrate...)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"pg_dsn": config.PG_DSN,
		}).WithError(err).Fatal("failed to AutoMigrate")
	}
}

func DbLoadSettingAsUint64(key string, defaultValue uint64) (uint64, error) {
	setting := &model.DbSetting{}

	err := db.Take(setting, "key = ?", key).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return defaultValue, err
	}

	if setting.Value == "" {
		return defaultValue, nil
	}

	result, err := strconv.ParseUint(setting.Value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("db setting key=%s value=%s parse error=\"%w\"", key, setting.Value, err)
	}

	return result, nil
}

func DbStoreSettingAsUint64(key string, value uint64) error {
	setting := &model.DbSetting{
		Key:   key,
		Value: strconv.FormatUint(value, 10),
	}

	return db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(setting).Error
}

func DbLoad(inOut model.GormLoad, key interface{}) error {
	return db.Take(inOut, inOut.GetGormDefaultCondition(), key).Error
}

func DbStore(data interface{}) error {
	return db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(data).Error
}
