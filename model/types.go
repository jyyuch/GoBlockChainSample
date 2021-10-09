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
	TranxHash []string `json:"transactions"`
}

type Tranx struct {
	Hash  string      `json:"tx_hash"`
	From  string      `json:"from"`
	To    string      `json:"to"`
	Nonce uint64      `json:"nonce"`
	Data  string      `json:"data"`
	Value string      `json:"value"`
	Logs  []*TranxLog `json:"logs"`
}

type TranxLog struct {
	Index uint   `json:"index"`
	Data  string `json:"data"`
}
