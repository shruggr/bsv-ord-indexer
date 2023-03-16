package lib

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/GorillaPool/go-junglebus"
	"github.com/GorillaPool/go-junglebus/models"
	"github.com/joho/godotenv"
	"github.com/libsv/go-bt/v2"
)

var txCache = make(map[string]*bt.Tx)
var jbClient *junglebus.JungleBusClient

func init() {
	godotenv.Load("../.env")
	var err error
	jb := os.Getenv("JUNGLEBUS")
	if jb == "" {
		jb = "https://junglebus.gorillapool.io"
	}
	jbClient, err = junglebus.New(
		junglebus.WithHTTP(jb),
	)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func LoadTx(txid string) (*bt.Tx, error) {
	if tx, ok := txCache[txid]; ok {
		return tx, nil
	}

	resp, err := http.Get(fmt.Sprintf("https://junglebus.gorillapool.io/v1/transaction/get/%s/bin", txid))
	if err != nil {
		return nil, err
	}

	rawtx, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	tx, err := bt.NewTxFromBytes(rawtx)
	if err != nil {
		return nil, err
	}

	txCache[txid] = tx
	return tx, nil
}

func LoadTxData(txid string) (*models.Transaction, error) {
	return jbClient.GetTransaction(context.Background(), txid)
}
