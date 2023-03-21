package bsvordindexer

import (
	"encoding/binary"
	"fmt"

	"github.com/libsv/go-bt/v2"
)

func LoadOrigin(txid []byte, vout uint32, maxDepth uint32) (origin []byte, err error) {
	outpoint := binary.BigEndian.AppendUint32(txid, vout)
	rows, err := Db.Query(`SELECT origin
		FROM ordinals
		WHERE outpoint=$1 AND outsat=$2 AND origin IS NOT NULL`,
		outpoint,
		0,
	)
	if err != nil {
		return
	}
	if rows.Next() {
		err = rows.Scan(&origin)
		rows.Close()
		return
	}
	rows.Close()

	if maxDepth == 0 {
		return
	}
	fmt.Printf("Indexing Origin %x %d\n", txid, vout)

	txData, err := LoadTxData(txid)
	if err != nil {
		return
	}
	tx, err := bt.NewTxFromBytes(txData.Transaction)
	if err != nil {
		return
	}
	_, err = ProcessInsTx(tx, txData.BlockHeight, uint32(txData.BlockIndex))
	if err != nil {
		return nil, err
	}

	if int(vout) >= len(tx.Outputs) {
		err = &HttpError{
			StatusCode: 400,
			Err:        fmt.Errorf("vout out of range"),
		}
		return
	}
	if tx.Outputs[vout].Satoshis != 1 {
		err = &HttpError{
			StatusCode: 400,
			Err:        fmt.Errorf("vout %d is not 1 satoshi", vout),
		}
		return
	}

	var txOutSat uint64
	for _, out := range tx.Outputs[0:vout] {
		txOutSat += out.Satoshis
	}

	var inSats uint64
	for _, input := range tx.Inputs {
		inTx, err := LoadTx(input.PreviousTxID())
		if err != nil {
			return nil, err
		}
		out := inTx.Outputs[input.PreviousTxOutIndex]

		if inSats+out.Satoshis < txOutSat {
			inSats += out.Satoshis
			continue
		}

		if inTx.Outputs[input.PreviousTxOutIndex].Satoshis > 1 {
			origin = outpoint
		} else {
			fmt.Printf("Loading Parent Origin %x %d\n", txid, vout)
			origin, err = LoadOrigin(input.PreviousTxID(), input.PreviousTxOutIndex, maxDepth-1)
			if err != nil {
				return nil, err
			}
		}

		if len(origin) > 0 {
			_, err = Db.Exec(`INSERT INTO ordinals(outpoint, outsat, origin)
				VALUES($1, 0, $2)
				ON CONFLICT(outpoint, outsat) DO UPDATE
					SET origin=EXCLUDED.origin`,
				outpoint,
				origin,
			)
			if err != nil {
				return nil, err
			}
		}
		break
	}
	return origin, nil
}
