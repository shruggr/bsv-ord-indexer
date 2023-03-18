package bsvordindexer

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/GorillaPool/go-junglebus"
	"github.com/GorillaPool/go-junglebus/models"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/libsv/go-bt/v2"
)

// var txCache = make(map[string]*bt.Tx)
var Db *sql.DB
var JBClient *junglebus.JungleBusClient

func init() {
	godotenv.Load("../.env")
	var err error
	Db, err = sql.Open("postgres", os.Getenv("POSTGRES"))
	if err != nil {
		log.Fatal(err)
	}
	jb := os.Getenv("JUNGLEBUS")
	if jb == "" {
		jb = "https://junglebus.gorillapool.io"
	}
	JBClient, err = junglebus.New(
		junglebus.WithHTTP(jb),
	)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func LoadTx(txid string) (tx *bt.Tx, err error) {
	resp, err := http.Get(fmt.Sprintf("https://junglebus.gorillapool.io/v1/transaction/get/%s/bin", txid))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		err = &HttpError{
			StatusCode: resp.StatusCode,
			Err:        err,
		}
		return
	}
	rawtx, err := io.ReadAll(resp.Body)
	fmt.Println("RAWTX:", len(rawtx))
	if err != nil {
		return nil, err
	}
	return bt.NewTxFromBytes(rawtx)
}

func LoadTxData(txid string) (*models.Transaction, error) {
	return JBClient.GetTransaction(context.Background(), txid)
}

// ByteString is a byte array that serializes to hex
type ByteString []byte

// MarshalJSON serializes ByteArray to hex
func (s ByteString) MarshalJSON() ([]byte, error) {
	bytes, err := json.Marshal(fmt.Sprintf("%x", string(s)))
	return bytes, err
}

// UnmarshalJSON deserializes ByteArray to hex
func (s *ByteString) UnmarshalJSON(data []byte) error {
	var x string
	err := json.Unmarshal(data, &x)
	if err == nil {
		str, e := hex.DecodeString(x)
		*s = ByteString([]byte(str))
		err = e
	}

	return err
}
