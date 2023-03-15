package main

import (
	"os"
	"strconv"

	"github.com/libsv/go-bt/v2"
	"github.com/shruggr/bsv-ord-indexer/lib"
)

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

	loadOrdinal(txid, uint32(vout), sat)
}

func loadOrdinal(txid string, vout uint32, sat uint64) (uint64, error) {
	var satoshi *Satoshi
	// TODO: Load satoshi from database.
	if satoshi != nil {
		return satoshi.OrdID, nil
	}

	tx, err := lib.LoadTx(txid)
	if err != nil {
		return 0, err
	}
	var txSat uint64
	for _, out := range tx.Outputs[0:vout] {
		txSat += out.Satoshis
	}

	var inSats uint64
	for _, input := range tx.Inputs {
		inTx, err := lib.LoadTx(input.PreviousTxIDStr())
		if err != nil {
			return 0, err
		}
		out := inTx.Outputs[input.PreviousTxOutIndex]

		if inSats+out.Satoshis < txSat {
			inSats += out.Satoshis
			continue
		}

		sat = txSat - inSats
		if inTx.IsCoinbase() {
			var height uint32
			txData, err := lib.LoadTxData(txid)
			if err != nil {
				return 0, err
			}
			height = txData.BlockHeight
			satoshis := subsidy(height)
			if sat < satoshis {
				satoshi.OrdID = firstOrdinal(height) + sat
			} else {
				var input *bt.Input
				for satoshis < sat {
					// TODO: find input sat

				}
				satoshi.OrdID, err = loadOrdinal(input.PreviousTxIDStr(), input.PreviousTxOutIndex, sat)
				if err != nil {
					return 0, err
				}
			}
		} else {
			satoshi.OrdID, err = loadOrdinal(input.PreviousTxIDStr(), input.PreviousTxOutIndex, sat)
			if err != nil {
				return 0, err
			}
		}

		// TODO: Save OrdId to database.
	}
	return satoshi.OrdID, nil
}

func subsidy(height uint32) uint64 {
	return 50 * 100_000_000 >> height / 210_000
}

func firstOrdinal(height uint32) uint64 {
	start := uint64(0)
	for i := uint32(0); i < height; i++ {
		start += subsidy(i)
	}
	return start
}

// # assign ordinals in given block
// def assign_ordinals(block):
//   first = first_ordinal(block.height)
//   last = first + subsidy(block.height)
//   coinbase_ordinals = list(range(first, last))

//   for transaction in block.transactions[1:]:
//     ordinals = []
//     for input in transaction.inputs:
//       ordinals.extend(input.ordinals)

//     for output in transaction.outputs:
//       output.ordinals = ordinals[:output.value]
//       del ordinals[:output.value]

//     coinbase_ordinals.extend(ordinals)

//   for output in block.transaction[0].outputs:
//     output.ordinals = coinbase_ordinals[:output.value]
//     del coinbase_ordinals[:output.value]

// Satoshi Indexer
// ===============
// 1. Load transaction data.
// 2. Find output satoshi index by counting satoshis in previous outputs.
// 3. Find corresponding input and determine index within input txo.
// 4. Load input transaction data.
// 5. If input transaction is not a coinbase, then repeat steps 1-4.
// 6. If input transaction coinbase and input txo index < block subsidy amount, calculate and return ordinal id.
// 7. If input transaction coinbase and input txo index >= subsidy amount:
// 8.   Load block data.

// 9.   Load each transaction to determine fees
