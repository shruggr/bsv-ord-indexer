package bsvordindexer

import (
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
)

var getOrigin *sql.Stmt
var setOrigin *sql.Stmt

func init() {
	var err error
	getOrigin, err = Db.Prepare(`SELECT origin
		FROM ordinals
		WHERE outpoint=$1 AND outsat=$2 AND origin IS NOT NULL`,
	)
	if err != nil {
		log.Fatal(err)
	}

	setOrigin, err = Db.Prepare(`INSERT INTO ordinals(outpoint, outsat, origin)
		VALUES($1, 0, $2)
		ON CONFLICT(outpoint, outsat) DO UPDATE
			SET origin=$2`,
	)
	if err != nil {
		log.Fatal(err)
	}
}

func LoadOrigin(txid string, vout uint32) (origin []byte, err error) {
	outpoint, err := hex.DecodeString(txid)
	if err != nil {
		return
	}
	outpoint = binary.BigEndian.AppendUint32(outpoint, vout)
	rows, err := getOrigin.Query(outpoint, 0)
	if err != nil {
		return
	}
	if rows.Next() {
		err = rows.Scan(&origin)
		return
	}

	fmt.Printf("Indexing %x\n", outpoint)

	tx, err := LoadTx(txid)
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
		inTx, err := LoadTx(input.PreviousTxIDStr())
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
			origin, err = LoadOrigin(input.PreviousTxIDStr(), input.PreviousTxOutIndex)
			if err != nil {
				return nil, err
			}
		}

		_, err = setOrigin.Exec(outpoint, origin)
		if err != nil {
			return nil, err
		}
		break
	}
	return origin, nil
}
