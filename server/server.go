package main

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/shruggr/bsv-ord-indexer/lib"
	"github.com/shruggr/bsv-ord-indexer/origin"
)

func main() {
	r := gin.Default()

	r.GET("/api/origin/:txid/:vout", func(c *gin.Context) {
		txid := c.Param("txid")
		vout, err := strconv.ParseUint(c.Param("vout"), 10, 32)
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("error: %s", err))
			return
		}

		origin, err := origin.LoadOrgin(txid, uint32(vout))
		if err != nil {
			if httpErr, ok := err.(*lib.HttpError); ok {
				c.String(httpErr.StatusCode, "%v", httpErr.Err)
			} else {
				c.String(http.StatusInternalServerError, "%v", err)
			}
			return
		}
		c.Header("cache-control", "max-age=604800,immutable")
		c.String(http.StatusOK, hex.EncodeToString(origin))
	})

	r.POST("/api/inscription/:txid", func(c *gin.Context) {
		tx, err := lib.LoadTx(c.Param("txid"))
		if err != nil {
			handleError(c, err)
		}
		err = lib.ProcessInsTx(tx, 0, 0)
		if err != nil {
			handleError(c, err)
		}
	})

	r.GET("/api/file/origin/:origin", func(c *gin.Context) {
		origin, err := hex.DecodeString(c.Param("origin"))
		if err != nil {
			handleError(c, err)
			return
		}
		ins, err := lib.LoadInsByOrigin(origin)
		if err != nil {
			handleError(c, err)
			return
		}
		c.Header("cache-control", "max-age=604800,immutable")
		c.Data(http.StatusOK, ins.Type, ins.Body)
	})

	r.GET("/api/file/ordinal/:ordinal", func(c *gin.Context) {
		origin, err := hex.DecodeString(c.Param("origin"))
		if err != nil {
			handleError(c, err)
			return
		}
		ins, err := lib.LoadInsByOrigin(origin)
		if err != nil {
			handleError(c, err)
			return
		}
		c.Header("cache-control", "max-age=604800,immutable")
		c.Data(http.StatusOK, ins.Type, ins.Body)
	})

	r.GET("/api/file/hash/:hash", func(c *gin.Context) {
		origin, err := hex.DecodeString(c.Param("origin"))
		if err != nil {
			handleError(c, err)
			return
		}
		ins, err := lib.LoadInsByOrigin(origin)
		if err != nil {
			handleError(c, err)
			return
		}
		c.Header("cache-control", "max-age=604800,immutable")
		c.Data(http.StatusOK, ins.Type, ins.Body)
	})

	r.GET("/api/file/inscription/:id", func(c *gin.Context) {
		origin, err := hex.DecodeString(c.Param("origin"))
		if err != nil {
			handleError(c, err)
			return
		}
		ins, err := lib.LoadInsByOrigin(origin)
		if err != nil {
			handleError(c, err)
			return
		}
		c.Header("cache-control", "max-age=604800,immutable")
		c.Data(http.StatusOK, ins.Type, ins.Body)
	})

	listen := os.Getenv("LISTEN")
	if listen == "" {
		listen = "0.0.0.0:8080"
	}
	r.Run(listen) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func handleError(c *gin.Context, err error) {
	if httpErr, ok := err.(*lib.HttpError); ok {
		c.String(httpErr.StatusCode, "%v", httpErr.Err)
	} else {
		c.String(http.StatusInternalServerError, "%v", err)
	}
}
