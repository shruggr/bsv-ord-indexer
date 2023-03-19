package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/GorillaPool/go-junglebus"
	jbModels "github.com/GorillaPool/go-junglebus/models"
	"github.com/joho/godotenv"
	"github.com/libsv/go-bt/v2"
	bsvord "github.com/shruggr/bsv-ord-indexer"
)

const INDEXER = "ord"

var db *sql.DB

func init() {
	godotenv.Load("../.env")

	var err error
	db, err = sql.Open("postgres", os.Getenv("POSTGRES"))
	if err != nil {
		log.Fatal(err)
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

	wg.Add(1)

	var fromBlock uint32
	row := db.QueryRow(`SELECT height
			FROM progress
			WHERE indexer='ord'`,
	)
	row.Scan(&fromBlock)
	if fromBlock < bsvord.TRIGGER {
		fromBlock = bsvord.TRIGGER
	}

	if _, err = junglebusClient.Subscribe(
		context.Background(),
		os.Getenv("ORD"),
		uint64(fromBlock),
		junglebus.EventHandler{
			OnTransaction: onOrdHandler,
			OnMempool:     onOrdHandler,
			OnStatus: func(status *jbModels.ControlResponse) {
				log.Printf("[STATUS]: %v\n", status)
				if status.StatusCode == 200 {
					if _, err := db.Exec(`INSERT INTO progress(indexer, height)
						VALUES($1, $2)
						ON CONFLICT(indexer) DO UPDATE
							SET height=$2`,
						INDEXER,
						status.Block,
					); err != nil {
						log.Print(err)
					}
				}
			},
			OnError: func(err error) {
				log.Printf("[ERROR]: %v", err)
			},
		},
	); err != nil {
		log.Printf("ERROR: failed getting subscription %s", err.Error())
		wg.Done()
	}

	wg.Wait()
}

func onOrdHandler(txResp *jbModels.TransactionResponse) {
	fmt.Printf("[TX]: %d: %v\n", txResp.BlockHeight, txResp.Id)
	tx, err := bt.NewTxFromBytes(txResp.Transaction)
	if err != nil {
		log.Printf("OnTransaction Parse Error: %s %+v\n", txResp.Id, err)

	}
	_, err = bsvord.ProcessInsTx(tx, txResp.BlockHeight, uint32(txResp.BlockIndex))
	if err != nil {
		log.Printf("OnTransaction Ins Error: %s %+v\n", txResp.Id, err)
	}
}
