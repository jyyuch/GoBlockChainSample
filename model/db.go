package model

type GormLoad interface {
	GetGormDefaultCondition() string
}

type DbSetting struct {
	Key   string `gorm:"primarykey"`
	Value string `gorm:"not null"`
}

type DbBlockBase struct {
	Num        uint64 `json:"block_num" gorm:"primarykey;autoIncrement:false"`
	Hash       string `json:"block_hash" gorm:"uniqueIndex;not null"`
	Time       uint64 `json:"block_time" gorm:"not null"`
	ParentHash string `json:"parent_hash" gorm:"uniqueIndex;not null"`
}

type DbBlock struct {
	DbBlockBase
	Transactions []DbTranx `json:"transactions" gorm:"foreignKey:BlockNum;references:Num"`
}

type DbTranx struct {
	Hash      string       `json:"tx_hash" gorm:"primarykey"`
	BlockNum  uint64       `json:"_" gorm:"index;not null"`
	BlockHash string       `json:"_" gorm:"index;not null"`
	From      string       `json:"from" gorm:"not null"`
	To        string       `json:"to" gorm:"not null"`
	Nonce     uint64       `json:"nonce" gorm:"not null"`
	Data      string       `json:"data"`
	Value     string       `json:"value"`
	Logs      []DbTranxLog `json:"logs" gorm:"foreignKey:TranxHash;references:Hash"`
}

type DbTranxLog struct {
	TranxHash string `json:"_" gorm:"primarykey"`
	Index     uint   `json:"index" gorm:"primarykey"`
	Data      string `json:"data"`
}

var TablesMigrate = []interface{}{
	&DbSetting{},
	&DbBlock{},
	&DbTranx{},
	&DbTranxLog{},
}

func (it *DbSetting) GetGormDefaultCondition() string {
	return "key = ?"
}

func (it *DbBlock) GetGormDefaultCondition() string {
	return "num = ?"
}

func (it *DbTranx) GetGormDefaultCondition() string {
	return "hash = ?"
}
