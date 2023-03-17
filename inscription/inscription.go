package main

import (
	"bytes"
	"context"
	"database/sql"
	"log"
	"os"
	"sync"
	"time"

	"github.com/GorillaPool/go-junglebus"
	jbModels "github.com/GorillaPool/go-junglebus/models"
	"github.com/joho/godotenv"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
)

const TRIGGER = 784000 // Placeholder
var fromBlock uint64
var db *sql.DB
var insInscription *sql.Stmt

var ordPattern []byte

func init() {
	godotenv.Load("../.env")

	var err error
	db, err = sql.Open("postgres", os.Getenv("POSTGRES"))
	if err != nil {
		log.Fatal(err)
	}

	insInscription, err := db.Prepare(`INSERT INTO inscriptions(txid, vout, sat, height, idx)
		VALUES($1, $2, $3, $4, $5)
		ON CONFLICT(txid, vout, sat) DO NOTHING`,
	)
	if err != nil {
		log.Fatal(err)
	}

	row := db.QueryRow(`SELECT height
		FROM progress
		WHERE indexer='inscriptions'`,
	)
	row.Scan(&fromBlock)
	if fromBlock < TRIGGER {
		fromBlock = TRIGGER
	}

	ordScript, err := bscript.NewFromASM("OP_FALSE OP_IF 6f7264")
	ordPattern = []byte(*ordScript)
}

func main() {
	var wg sync.WaitGroup
	junglebusClient, err := junglebus.New(
		junglebus.WithHTTP("https://junglebus.gorillapool.io"),
	)
	if err != nil {
		log.Fatalln(err.Error())
	}

	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) < 2 {
		panic("no subscription id or block height given")
	}
	subscriptionID := argsWithoutProg[0]

	eventHandler := junglebus.EventHandler{
		// do not set this function to leave out mined transactions
		OnTransaction: func(tx *jbModels.TransactionResponse) {
			log.Printf("[TX]: %d: %v", tx.BlockHeight, tx.Id)
		},
		// do not set this function to leave out mempool transactions
		OnMempool: func(tx *jbModels.TransactionResponse) {
			log.Printf("[MEMPOOL TX]: %v", tx.Id)
		},
		OnStatus: func(status *jbModels.ControlResponse) {
			log.Printf("[STATUS]: %v", status)
		},
		OnError: func(err error) {
			log.Printf("[ERROR]: %v", err)
		},
	}

	wg.Add(1)
	var subscription *junglebus.Subscription
	if subscription, err = junglebusClient.Subscribe(context.Background(), subscriptionID, fromBlock, eventHandler); err != nil {
		log.Printf("ERROR: failed getting subscription %s", err.Error())
	} else {
		time.Sleep(10 * time.Second) // stop after 10 seconds
		if err = subscription.Unsubscribe(); err != nil {
			log.Printf("ERROR: failed unsubscribing %s", err.Error())
		}
		wg.Done()
	}

	wg.Wait()
}

func processInscription(txResp *jbModels.TransactionResponse) (err error) {
	tx, err := bt.NewTxFromBytes(txResp.Transaction)
	if err != nil {
		return
	}

	for vout, txout := range tx.Outputs {
		idx := bytes.Index(*txout.LockingScript, ordPattern)
		if idx == -1 {
			continue
		}

		_, err = insInscription.Exec(tx.TxIDBytes(),
			vout,
			0,
			txResp.BlockHeight,
			txResp.BlockIndex,
		)
	}
}
