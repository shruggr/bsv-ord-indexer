package models

import (
	"github.com/bitcoinschema/go-b"
	magic "github.com/bitcoinschema/go-map"
)

type Satoshi struct {
	Txid   string
	Vout   uint32
	OutSat uint64
	Origin *Satoshi
	OrdID  uint64
}

type Inscription struct {
	ID      uint64
	Txid    string
	Vout    uint32
	Satoshi uint64
	File    *b.B
	Map     magic.MAP
	OrdId   uint64
}
