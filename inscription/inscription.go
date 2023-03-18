package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/GorillaPool/go-junglebus"
	jbModels "github.com/GorillaPool/go-junglebus/models"
	"github.com/joho/godotenv"
	"github.com/libsv/go-bt/v2"
	"github.com/shruggr/bsv-ord-indexer/lib"
)

const TRIGGER = 784000 // Placeholder
const INDEXER = "inscriptions"

// const
var fromBlock uint64
var db *sql.DB
var setProgress *sql.Stmt

func init() {
	godotenv.Load("../.env")

	var err error
	db, err = sql.Open("postgres", os.Getenv("POSTGRES"))
	if err != nil {
		log.Fatal(err)
	}

	setProgress, err = db.Prepare(`INSERT INTO progress(indexer, height)
		VALUES($1, $2)
		ON CONFLICT(indexer) DO UPDATE
			SET height=$2`,
	)
	if err != nil {
		log.Fatal(err)
	}

	row := db.QueryRow(`SELECT height
		FROM progress
		WHERE indexer=$1`,
		INDEXER,
	)
	row.Scan(&fromBlock)
	if fromBlock < TRIGGER {
		fromBlock = TRIGGER
	}
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
		OnTransaction: func(txResp *jbModels.TransactionResponse) {
			fmt.Printf("[TX]: %d: %v", txResp.BlockHeight, txResp.Id)
			tx, err := bt.NewTxFromBytes(txResp.Transaction)
			if err != nil {
				return
			}
			err = lib.ProcessInsTx(tx, txResp.BlockHeight, uint32(txResp.BlockIndex))
			if err != nil {
				log.Printf("OnTransaction Error: %x %+v", txResp.Id, err)
			}
		},
		// do not set this function to leave out mempool transactions
		// OnMempool: func(tx *jbModels.TransactionResponse) {
		// 	fmt.Printf("[MEMPOOL TX]: %v", tx.Id)
		// 	err = processInscription(tx)
		// 	if err != nil {
		// 		log.Printf("OnMempool Error: %x %+v", tx.Id, err)
		// 	}
		// },
		OnStatus: func(status *jbModels.ControlResponse) {
			log.Printf("[STATUS]: %v", status)
			if status.StatusCode == 200 {
				if _, err := setProgress.Exec(); err != nil {
					log.Print(err)
				}
			}
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
