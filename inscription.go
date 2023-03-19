package bsvordindexer

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
)

var PATTERN []byte

func init() {
	val, err := hex.DecodeString("0063036f7264")
	if err != nil {
		log.Panic(err)
	}
	PATTERN = val
}

type Inscription struct {
	Body []byte
	Type string
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

type File struct {
	Hash ByteString `json:"hash"`
	Size uint32     `json:"size"`
	Type string     `json:"type"`
}

type InscriptionMeta struct {
	Id      uint64     `json:"id"`
	Txid    ByteString `json:"txid"`
	Vout    uint32     `json:"vout"`
	File    File       `json:"file"`
	Origin  ByteString `json:"origin"`
	Ordinal uint32     `json:"ordinal"`
	Height  uint32     `json:"height"`
	Idx     uint32     `json:"idx"`
}

func (im *InscriptionMeta) Save() (err error) {
	_, err = db.Exec(`
		INSERT INTO inscriptions(txid, vout, height, idx, filehash, filesize, filetype, origin)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT(txid, vout) DO UPDATE
			SET height=EXCLUDED.height, idx=EXCLUDED.idx`,
		im.Txid,
		im.Vout,
		im.Height,
		im.Idx,
		im.File.Hash,
		im.File.Size,
		im.File.Type,
		im.Origin,
	)
	return
}

func ProcessInsTx(tx *bt.Tx, height uint32, idx uint32) (ins []*InscriptionMeta, err error) {
	for vout, txout := range tx.Outputs {
		inscription := InscriptionFromScript(*txout.LockingScript)
		if inscription == nil {
			continue
		}

		hash := sha256.Sum256(inscription.Body)

		im := &InscriptionMeta{
			Txid: tx.TxIDBytes(),
			Vout: uint32(vout),
			File: File{
				Hash: hash[:],
				Size: uint32(len(inscription.Body)),
				Type: inscription.Type,
			},
			Height: height,
			Idx:    idx,
		}
		im.Origin, err = LoadOrigin(tx.TxID(), uint32(vout), 100)
		if err != nil {
			return
		}
		err = im.Save()
		if err != nil {
			return
		}
		ins = append(ins, im)
	}

	return
}

func GetInsByOrigin(origin []byte) (ins []*InscriptionMeta, err error) {
	rows, err := db.Query(`SELECT txid, vout, filehash, filesize, filetype, id, origin, ordinal, height, idx
		FROM inscriptions
		WHERE origin=$1
		ORDER BY height DESC, idx DESC`,
		origin,
	)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		im := &InscriptionMeta{}
		err = rows.Scan(
			&im.Txid,
			&im.Vout,
			&im.File.Hash,
			&im.File.Size,
			&im.File.Type,
			&im.Id,
			&im.Origin,
			&im.Ordinal,
			&im.Height,
			&im.Idx,
		)
		if err != nil {
			return
		}
		ins = append(ins, im)
	}
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
	defer rows.Close()

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
