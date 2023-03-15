package models

import (
	"fmt"

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

func (s *Satoshi) OriginString() string {
	if s.Origin == nil {
		return fmt.Sprintf("%s:%d:%d", s.Txid, s.Vout, s.OutSat)
	}
	return fmt.Sprintf("%s:%d:%d", s.Origin.Txid, s.Origin.Vout, s.Origin.OutSat)
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
