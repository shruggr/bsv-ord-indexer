package lib

type Block struct {
	Hash    string   `json:"hash"`
	Height  uint     `json:"height"`
	TxCount uint     `json:"txcount"`
	Tx      []string `json:"tx"`
	Pages   []string `json:"pages"`
}
