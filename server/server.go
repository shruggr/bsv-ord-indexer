package main

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/shruggr/bsv-ord-indexer/origin"
)

func main() {
	r := gin.Default()

	r.GET("/api/origin/:txid/:vout", func(c *gin.Context) {
		c.Header("cache-control", "max-age=604800, immutable")
		txid := c.Param("txid")
		vout, err := strconv.ParseUint(c.Param("vout"), 10, 32)
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("error: %s", err))
			return
		}

		origin, err := origin.LoadOrgin(txid, uint32(vout))
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("error: %s", err))
			return
		}
		c.String(http.StatusOK, hex.EncodeToString(origin))
	})
	listen := os.Getenv("LISTEN")
	if listen == "" {
		listen = "0.0.0.0:8080"
	}
	r.Run(listen) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
