package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"sync"

	"github.com/GorillaPool/go-junglebus"
	jbModels "github.com/GorillaPool/go-junglebus/models"
	"github.com/joho/godotenv"
	"github.com/libsv/go-bt/v2"
	bsvord "github.com/shruggr/bsv-ord-indexer"
)

const TRIGGER = 782000 // Placeholder 783968
const INDEXER = "1sat"
const THREADS = 16

var db *sql.DB
var threadsChan = make(chan struct{}, THREADS)

func init() {
	godotenv.Load("../.env")

	var err error
	db, err = sql.Open("postgres", os.Getenv("POSTGRES"))
	if err != nil {
		log.Fatal(err)
	}
}

var wg sync.WaitGroup

func main() {
	junglebusClient, err := junglebus.New(
		junglebus.WithHTTP("https://junglebus.gorillapool.io"),
	)
	if err != nil {
		log.Fatalln(err.Error())
	}

	var fromBlock uint64
	row := db.QueryRow(`SELECT height
			FROM progress
			WHERE indexer='1sat'`,
	)
	row.Scan(&fromBlock)
	if fromBlock < TRIGGER {
		fromBlock = TRIGGER
	}

	var wg2 sync.WaitGroup
	wg2.Add(1)
	if _, err = junglebusClient.Subscribe(
		context.Background(),
		os.Getenv("ONESAT"),
		fromBlock,
		junglebus.EventHandler{
			OnTransaction: onOneSatHandler,
			OnMempool:     onOneSatHandler,
			OnStatus: func(status *jbModels.ControlResponse) {
				wg.Wait()
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
		wg2.Done()
	}

	wg2.Wait()
}

func onOneSatHandler(txResp *jbModels.TransactionResponse) {
	// fmt.Printf("[TX]: %d: %v\n", txResp.BlockHeight, txResp.Id)
	tx, err := bt.NewTxFromBytes(txResp.Transaction)
	if err != nil {
		log.Printf("OnTransaction Parse Error: %s %+v\n", txResp.Id, err)
	}
	height := uint32(txResp.BlockHeight)
	idx := uint32(txResp.BlockIndex)

	wg.Add(1)
	threadsChan <- struct{}{}
	go func() {
		for vout, txout := range tx.Outputs {
			if txout.Satoshis == 1 {
				_, err = bsvord.LoadOrigin(tx.TxID(), uint32(vout), 25)
				if err != nil {
					log.Printf("LoadOrigin Error: %s %+v\n", tx.TxID(), err)
				}
			}
		}
		_, err = bsvord.ProcessInsTx(tx, height, idx)
		if err != nil {
			log.Printf("ProcessInsTx Error: %s %+v\n", txResp.Id, err)
		}
		wg.Done()
		<-threadsChan
	}()

}
