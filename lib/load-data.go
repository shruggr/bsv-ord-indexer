package lib

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/GorillaPool/go-junglebus"
	"github.com/GorillaPool/go-junglebus/models"
	"github.com/libsv/go-bt/v2"
)

var txCache = make(map[string]*bt.Tx)
var jbClient *junglebus.JungleBusClient

func init() {
	var err error
	jbClient, err = junglebus.New(
		junglebus.WithHTTP("https://junglebus.gorillapool.io"),
	)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func LoadRawtx(txid string) (*bt.Tx, error) {
	if tx, ok := txCache[txid]; ok {
		return tx, nil
	}

	resp, err := http.Get(fmt.Sprintf("https://junglebus.gorillapool.io/v1/transaction/get/%s/bin", txid))
	if err != nil {
		return nil, err
	}

	rawtx := make([]byte, resp.ContentLength)
	_, err = resp.Body.Read(rawtx)
	if err != nil {
		return nil, err
	}
	tx, err := bt.NewTxFromBytes(rawtx)
	if err != nil {
		panic(err)
	}

	txCache[txid] = tx
	return tx, nil
}

func LoadTxData(txid string) (*models.Transaction, error) {
	return jbClient.GetTransaction(context.Background(), txid)
}

func LoadBlock(hash string) {

}
