package model

type BlockBase struct {
	Num        uint64 `json:"block_num"`
	Hash       string `json:"block_hash"`
	Time       uint64 `json:"block_time"`
	ParentHash string `json:"parent_hash"`
}

type ResponseBlocks struct {
	Blocks []*BlockBase `json:"blocks"`
}

type BlockTranx struct {
	BlockBase
	Tranx []string `json:"transactions"`
}
