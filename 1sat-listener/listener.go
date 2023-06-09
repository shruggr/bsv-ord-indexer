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

const INDEXER = "1sat"
const THREADS = 16

var db *sql.DB
var threadsChan = make(chan struct{}, THREADS)
var processedIdx = map[uint64]bool{}
var m sync.Mutex

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

	var fromBlock uint32
	row := db.QueryRow(`SELECT height+1
			FROM progress
			WHERE indexer='1sat'`,
	)
	row.Scan(&fromBlock)
	if fromBlock < bsvord.TRIGGER {
		fromBlock = bsvord.TRIGGER
	}

	var wg2 sync.WaitGroup
	wg2.Add(1)
	if _, err = junglebusClient.Subscribe(
		context.Background(),
		os.Getenv("ONESAT"),
		uint64(fromBlock),
		junglebus.EventHandler{
			OnTransaction: onOneSatHandler,
			// OnMempool:     onOneSatHandler,
			OnStatus: func(status *jbModels.ControlResponse) {
				log.Printf("[STATUS]: %v\n", status)
				if status.StatusCode == 200 {
					wg.Wait()
					if _, err := db.Exec(`INSERT INTO progress(indexer, height)
						VALUES($1, $2)
						ON CONFLICT(indexer) DO UPDATE
							SET height=$2`,
						INDEXER,
						status.Block,
					); err != nil {
						log.Print(err)
					}
					m.Lock()
					processedIdx = map[uint64]bool{}
					m.Unlock()
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
	fmt.Printf("[TX]: %d: %v\n", txResp.BlockHeight, txResp.Id)

	m.Lock()
	if _, ok := processedIdx[txResp.BlockIndex]; txResp.BlockHeight > 0 && ok {
		fmt.Println("Already Processed:", txResp.Id)
		return
	}
	m.Lock()
	tx, err := bt.NewTxFromBytes(txResp.Transaction)
	if err != nil {
		log.Printf("OnTransaction Parse Error: %s %+v\n", txResp.Id, err)
	}

	wg.Add(1)
	threadsChan <- struct{}{}
	txid := tx.TxIDBytes()
	go func(txResp *jbModels.TransactionResponse) {
		for vout, txout := range tx.Outputs {
			if txout.Satoshis == 1 {
				_, err = bsvord.LoadOrigin(txid, uint32(vout), 25)
				if err != nil {
					log.Printf("LoadOrigin Error: %s %+v\n", tx.TxID(), err)
				}
			}
		}
		wg.Done()
		if txResp.BlockHeight > 0 {
			m.Lock()
			processedIdx[txResp.BlockIndex] = true
			m.Unlock()
		}
		<-threadsChan
	}(txResp)

}
