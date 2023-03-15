package main

import (
	"encoding/json"
	"log"

	"github.com/shruggr/bsv-ord-indexer/lib"
)

func main() {
	data := `{ "hash": "000", "height": 0, "tx": ["1", "2" ], "txcount": 2, "pages": null }`
	block := &lib.Block{}
	err := json.Unmarshal([]byte(data), &block)
	if err != nil {
		panic(err)
	}
	log.Println("Block:", block)
}
