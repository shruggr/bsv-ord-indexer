package ordinal

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var db *sql.DB
var GetOrd *sql.Stmt
var SetOrd *sql.Stmt
var GetTxOuts *sql.Stmt
var GetTxIns *sql.Stmt
var GetBlkFeeTxIn *sql.Stmt

func init() {
	godotenv.Load("../.env")

	var err error
	db, err = sql.Open("postgres", os.Getenv("POSTGRES"))
	if err != nil {
		log.Fatal(err)
	}

	GetOrd, err = db.Prepare(`SELECT ordinal
		FROM ordinals
		WHERE outpoint = $1 AND outsat=$2 AND ordinal IS NOT NULL`,
	)
	if err != nil {
		log.Fatal(err)
	}

	SetOrd, err = db.Prepare(`INSERT INTO ordinals(outpoint, outsat, ordinal)
		VALUES($1, $2, $3)
		ON CONFLICT(outpoint, outsat) DO UPDATE
			SET ordinal=$2`,
	)
	if err != nil {
		log.Fatal(err)
	}

	GetTxOuts, err = db.Prepare(`SELECT vout, satoshis
		FROM txos
		WHERE txid=$1
		ORDER BY vout ASC`,
	)
	if err != nil {
		log.Fatal(err)
	}

	GetTxIns, err = db.Prepare(`SELECT vin, txid, vout, satoshis, coinbase
		FROM txos
		WHERE spend=$1
		ORDER BY vin ASC`,
	)
	if err != nil {
		log.Fatal(err)
	}

	GetBlkFeeTxIn, err = db.Prepare(`SELECT txid, fee, acc
		FROM blk_txns
		WHERE height=$1 AND acc<$2
		ORDER BY acc DESC
		LIMIT 1`,
	)
	if err != nil {
		log.Fatal(err)
	}
}

func LoadOrdinal(txid []byte, vout uint32, sat uint64) (ordinal uint64, err error) {
	outpoint := binary.BigEndian.AppendUint32(txid, vout)
	rows, err := GetOrd.Query(outpoint, sat)
	if err != nil {
		return
	}

	if rows.Next() {
		err = rows.Scan(&ordinal)
		return
	}

	var txOutSat uint64
	var i uint32
	rows, err = GetTxOuts.Query(txid, vout)
	if err != nil {
		return
	}
	for rows.Next() {
		var satoshis uint64
		err = rows.Scan(&i, &satoshis)
		if err != nil {
			return
		}
		if vout == i {
			txOutSat += sat
			break
		} else if vout < i {
			txOutSat += satoshis
		}
	}
	rows.Close()
	if i < vout {
		return 0, fmt.Errorf("vout out of range")
	}

	rows, err = GetTxIns.Query(txid)
	if err != nil {
		return
	}
	var inSats uint64
	for rows.Next() {
		var vin uint32
		var satoshis uint64
		var coinbase uint32
		err = rows.Scan(&vin, &txid, &vout, &satoshis, &coinbase)
		if err != nil {
			return
		}

		if inSats+satoshis >= txOutSat {
			sat := txOutSat - inSats
			if coinbase > 0 {
				sub := subsidy(coinbase)
				if sub > sat {
					ordinal = firstOrdinal(coinbase) + sat
				} else {
					row := GetBlkFeeTxIn.QueryRow(coinbase, sat)
					err = row.Scan(&txid, &vout)
					if err != nil {
						return
					}
					ordinal, err = LoadOrdinal(txid, vout, sat)
				}
			} else {
				ordinal, err = LoadOrdinal(txid, vout, sat)
			}
			break
		}
		inSats += satoshis
	}
	rows.Close()
	if ordinal > 0 {
		_, err = SetOrd.Exec(outpoint, sat, ordinal)
	}

	return
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
