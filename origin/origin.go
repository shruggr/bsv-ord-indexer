package origin

import (
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/shruggr/bsv-ord-indexer/lib"
)

var db *sql.DB
var getOrigin *sql.Stmt
var setOrigin *sql.Stmt

func init() {
	godotenv.Load("../.env")

	var err error
	// var err error
	db, err = sql.Open("postgres", os.Getenv("POSTGRES"))
	if err != nil {
		log.Fatal(err)
	}

	getOrigin, err = db.Prepare(`SELECT origin
		FROM ordinals
		WHERE outpoint = $1`,
	)
	if err != nil {
		log.Fatal(err)
	}

	setOrigin, err = db.Prepare(`INSERT INTO ordinals(outpoint, origin)
		VALUES($1, $2)
		ON CONFLICT(outpoint) DO UPDATE
			SET origin=$2`,
	)
	if err != nil {
		log.Fatal(err)
	}
}

func LoadOrgin(txid string, vout uint32) (origin []byte, err error) {
	outpoint, err := hex.DecodeString(txid)
	if err != nil {
		return
	}
	outpoint = binary.BigEndian.AppendUint32(outpoint, vout)
	rows, err := getOrigin.Query(outpoint)
	if err != nil {
		return
	}
	if rows.Next() {
		err = rows.Scan(&origin)
		return
	}

	fmt.Printf("Indexing %x\n", outpoint)

	tx, err := lib.LoadTx(txid)
	if err != nil {
		return nil, err
	}
	if int(vout) >= len(tx.Outputs) {
		return nil, fmt.Errorf("vout out of range")
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

		if inTx.Outputs[input.PreviousTxOutIndex].Satoshis > 1 {
			origin = outpoint
		} else {
			origin, err = LoadOrgin(input.PreviousTxIDStr(), input.PreviousTxOutIndex)
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
