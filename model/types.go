package model

type BlockTranx struct {
	DbBlockBase
	TranxHash []string `json:"transactions"`
}

type ResponseBlocks struct {
	Blocks []*DbBlockBase `json:"blocks"`
}
