package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/shruggr/bsv-ord-indexer/lib"
)

type Txo struct {
	Script   []byte
	Satoshis uint64
}

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

func main() {
	txid := os.Args[1]

	vout, err := strconv.ParseUint(os.Args[2], 10, 32)
	if err != nil {
		panic(err)
	}

	sat, err := strconv.ParseUint(os.Args[3], 10, 64)
	if err != nil {
		panic(err)
	}

	loadOrgin(txid, uint32(vout), sat)
	fmt.Println()
}

func loadOrgin(txid string, vout uint32, voutSat uint64) (*Satoshi, error) {
	var satoshi *Satoshi
	// TODO: Load satoshi from database.
	if satoshi != nil {
		return satoshi, nil
	}

	tx, err := lib.LoadTx(txid)
	if err != nil {
		return nil, err
	}
	if tx.Outputs[vout].Satoshis != 1 {
		return nil, fmt.Errorf("vout %d is not 1 satoshi", vout)
	}
	var txOutSat uint64
	for _, out := range tx.Outputs[0:vout] {
		txOutSat += out.Satoshis
	}

	var inSats uint64
	for _, input := range tx.Inputs {
		inTx, err := lib.LoadTx(input.PreviousTxIDStr())
		if err != nil {
			return nil, err
		}
		out := inTx.Outputs[input.PreviousTxOutIndex]

		if inSats+out.Satoshis < txOutSat {
			inSats += out.Satoshis
			continue
		}
		voutSat = txOutSat - inSats

		if inTx.Outputs[input.PreviousTxOutIndex].Satoshis > 1 {
			origin := &Satoshi{
				Txid:   txid,
				Vout:   vout,
				OutSat: voutSat,
				Origin: nil,
			}

			// Save to database
			return origin, nil
		}

		satoshi, err = loadOrgin(input.PreviousTxIDStr(), input.PreviousTxOutIndex, voutSat)
		if err != nil {
			return nil, err
		}
		break
	}
	return satoshi, nil
}
