package lib

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
)

var db *sql.DB

var PATTERN []byte

func init() {
	godotenv.Load("../.env")

	val, err := hex.DecodeString("0063036f7264")
	if err != nil {
		log.Panic(err)
	}
	PATTERN = val

	db, err = sql.Open("postgres", os.Getenv("POSTGRES"))
	if err != nil {
		log.Fatal(err)
	}
}

type Inscription struct {
	Body []byte
	Type string
}

type File struct {
	Hash []byte `json:`

}
type InscriptionMeta struct {
	Txid []byte
	Vout uint32

}

func InscriptionFromScript(lock []byte) (ins *Inscription) {
	idx := bytes.Index(lock, PATTERN)
	if idx == -1 {
		return
	}

	idx += len(PATTERN)
	if idx >= len(lock) {
		return
	}

	script := bscript.NewFromBytes((lock)[idx:])
	parts, err := bscript.DecodeParts(*script)
	if err != nil {
		return
	}

	ins = &Inscription{}
	for i := 0; i < len(parts); i++ {
		op := parts[i]
		if len(op) != 1 {
			break
		}
		opcode := op[0]
		switch opcode {
		case bscript.Op0:
			ins.Body = parts[i+1]
			return
		case bscript.Op1:
			ins.Type = string(parts[i+1])
		case bscript.OpENDIF:
			return
		}
		i++
	}
	return
}

func (ins *Inscription) Save(txid []byte, vout uint32, height uint32, idx uint32) (err error) {
	_, err = db.Exec(`
		INSERT INTO inscriptions(txid, vout, height, idx, filehash, filesize, filetype)
		VALUES($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT(txid, vout) DO UPDATE
			SET height=$4, idx=$5
			WHERE EXCLUDED.height > 0 AND EXCLUDED.idx > 0`,
		txid,
		vout,
		height,
		idx,
		sha256.Sum256(ins.Body),
		len(ins.Body),
		ins.Type,
	)
	return
}

func ProcessInsTx(tx *bt.Tx, height uint32, idx uint32) (err error) {
	for vout, txout := range tx.Outputs {
		inscription := InscriptionFromScript(*txout.LockingScript)
		if inscription == nil {
			continue
		}

		err = inscription.Save(
			tx.TxIDBytes(),
			uint32(vout),
			height,
			idx,
		)
		if err != nil {
			return
		}
	}

	return
}

func GetInsByOrigin(origin []byte) (ins []*Inscription, err error) {
	rows, err := db.Query(`SELECT txid, vout, filehash, filesize, filetype
		FROM inscriptions
		WHERE origin=$1
		ORDER BY height DESC, idx DESC`,
		origin,
	)
	if err != nil {
		return
	}

	for rows.Next() {
		ins := &Inscription{}
		err = rows.Scan(&ins.)
		if err != nil {
			return
		}
	}

	tx, err := LoadTx(hex.EncodeToString(txid))
	if err != nil {
		return
	}

	ins = InscriptionFromScript(*tx.Outputs[vout].LockingScript)
	return
}

func LoadInsByOrigin(origin []byte) (ins *Inscription, err error) {
	rows, err := db.Query(`SELECT txid, vout, filetype
		FROM inscriptions
		WHERE origin=$1
		ORDER BY height DESC, idx DESC
		LIMIT 1`,
		origin,
	)
	if err != nil {
		return
	}

	if !rows.Next() {
		err = &HttpError{
			StatusCode: 404,
			Err:        fmt.Errorf("not-found"),
		}
		return
	}
	var txid []byte
	var vout uint32
	var filetype string
	err = rows.Scan(&txid, &vout, &filetype)
	if err != nil {
		return
	}

	tx, err := LoadTx(hex.EncodeToString(txid))
	if err != nil {
		return
	}

	ins = InscriptionFromScript(*tx.Outputs[vout].LockingScript)
	return
}
