package main

import (
	"os"

	"github.com/bitcoinschema/go-b"
	"github.com/bitcoinschema/go-bob"
	"github.com/bitcoinschema/go-bpu"
	magic "github.com/bitcoinschema/go-map"

	"github.com/shruggr/bsv-ord-indexer/lib"
)

type Inscription struct {
	ID      uint64
	Txid    string
	Vout    uint32
	Satoshi uint64
	File    *b.B
	Map     magic.MAP
	OrdId   uint64
}

func main() {
	txid := os.Args[1]

	tx := lib.LoadRawtx(txid)

	bobTx, err := bob.NewFromTx(tx)
	if err != nil {
		panic(err)
	}
	var prevOut *bpu.Output
	var prevVout int
	for vout, out := range bobTx.Out {
		if *out.E.V > 0 {
			prevOut = &out
			prevVout = vout
			continue
		}
		var inscr *Inscription
		for _, tape := range out.Tape {
			if *tape.Cell[0].S == "ord" && prevOut != nil {
				inscr = &Inscription{
					Txid: tx.TxID(),
					Vout: uint32(prevVout),
				}
				continue
			}
			if *tape.Cell[0].S == b.Prefix {
				bOut, err := b.NewFromTape(tape)
				if err != nil {
					panic(err)
				}
				inscr.File = bOut
			}
			if *tape.Cell[0].S == magic.Prefix {
				mapOut, err := magic.NewFromTape(&tape)
				if err != nil {
					panic(err)
				}
				inscr.Map = mapOut
			}
		}
		if inscr != nil {
			// TODO: Save Inscription to database.
		}

		prevOut = nil
	}

}
