package bsvordindexer

import (
	"database/sql"
	"log"

	"github.com/libsv/go-bt/v2"
)

var insSpend *sql.Stmt
var insTxo *sql.Stmt

func init() {
	var err error
	insTxo, err = Db.Prepare(`INSERT INTO txos(txid, vout, satoshis, lockhash, origin)
		VALUES($1, $2, $3, $4)
		ON CONFLICT DO UPDATE SET 
			lockhash=EXCLUDED.lockhash, 
			satoshis=EXCLUDED.satoshis, 
			origin=EXCLUDED.origin
	`)
	if err != nil {
		log.Fatal(err)
	}

	insSpend, err = Db.Prepare(`INSERT INTO txos(txid, vout, spend, vin)
		VALUES($1, $2, $3, $4)
		ON CONFLICT DO UPDATE 
			SET spend=EXCLUDED.spend, vin=EXCLUDED.vin
	`)
	if err != nil {
		log.Fatal(err)
	}

}
func IndexTxn(tx *bt.Tx) (err error) {
	txid := tx.TxIDBytes()
	for vout, txout := range tx.Outputs {

	}
	for vin, txin := range tx.Inputs {
		_, err = insSpend.Exec(
			txin.PreviousTxID(),
			txin.PreviousTxOutIndex,
			txid,
			vin,
		)
	}
}
