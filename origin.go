package bsvordindexer

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/libsv/go-bt/v2"
)

var getOrigin *sql.Stmt
var insOrigin *sql.Stmt

func init() {
	var err error
	getOrigin, err = Db.Prepare(`SELECT origin
		FROM ordinals
		WHERE txid=$1 AND vout=$2 AND sat=$3`,
	)
	if err != nil {
		log.Fatal(err)
	}

	insOrigin, err = Db.Prepare(`INSERT INTO ordinals(txid, vout, sat, origin)
		VALUES($1, $2, 0, $3)
		ON CONFLICT(outpoint, outsat) DO UPDATE
			SET origin=EXCLUDED.origin`,
	)
	if err != nil {
		log.Fatal(err)
	}
}

func LoadOrigin(txid []byte, vout uint32, maxDepth uint32) (origin []byte, err error) {
	outpoint := binary.BigEndian.AppendUint32(txid, vout)
	rows, err := getOrigin.Query(txid, vout, 0)
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
		inTxData, err := LoadTxData(input.PreviousTxID())
		if err != nil {
			return nil, err
		}
		inTx, err := bt.NewTxFromBytes(inTxData.Transaction)
		if err != nil {
			return nil, err
		}
		out := inTx.Outputs[input.PreviousTxOutIndex]

		if inSats+out.Satoshis < txOutSat {
			inSats += out.Satoshis
			continue
		}

		if inTx.Outputs[input.PreviousTxOutIndex].Satoshis > 1 || inTxData.BlockHeight < TRIGGER {
			origin = outpoint
		} else {
			fmt.Printf("Loading Parent Origin %x %d\n", txid, vout)
			origin, err = LoadOrigin(input.PreviousTxID(), input.PreviousTxOutIndex, maxDepth-1)
			if err != nil {
				return nil, err
			}
		}

		if len(origin) > 0 {
			_, err = insOrigin.Exec(
				input.PreviousTxID(),
				input.PreviousTxOutIndex,
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
