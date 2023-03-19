package main

import (
	"log"
	"os"
	"strconv"

	"github.com/ordishs/go-bitcoin"
	bsvord "github.com/shruggr/bsv-ord-indexer"
)

var bit *bitcoin.Bitcoind

func init() {
	port, err := strconv.ParseInt(os.Getenv("BITCOIN_PORT"), 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	bit, err = bitcoin.New(
		os.Getenv("BITCOIN_HOST"),
		int(port),
		os.Getenv("BITCOIN_USER"),
		os.Getenv("BITCOIN_PASS"),
		false,
	)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	row := bsvord.Db.QueryRow(`
		SELECT hash, height
		FROM blocks
		WHERE status=2
		ORDER BY height DESC
		LIMIT 1
	`)
	var height int
	var iHash string
	err := row.Scan(&height, &iHash)
	if err != nil {
		log.Fatal(err)
	}
	


	block, err := bit.GetBlockByHeight(height)

	if

	if block.NextBlockHash != "" {
		block, err = bit.Get

	}

}


func processBlock(block)
