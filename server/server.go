package main

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	bsvord "github.com/shruggr/bsv-ord-indexer"
	_ "github.com/shruggr/bsv-ord-indexer/server/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	r := gin.Default()
	url := ginSwagger.URL("/swagger/doc.json")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	r.GET("/api/origins/:txid/:vout", func(c *gin.Context) {
		txid := c.Param("txid")
		vout, err := strconv.ParseUint(c.Param("vout"), 10, 32)
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("error: %s", err))
			return
		}

		origin, err := bsvord.LoadOrigin(txid, uint32(vout), 25)
		if err != nil {
			if httpErr, ok := err.(*bsvord.HttpError); ok {
				c.String(httpErr.StatusCode, "%v", httpErr.Err)
			} else {
				c.String(http.StatusInternalServerError, "%v", err)
			}
			return
		}
		c.Header("cache-control", "max-age=604800,immutable")
		c.String(http.StatusOK, hex.EncodeToString(origin))
	})

	r.POST("/api/inscriptions/:txid", func(c *gin.Context) {
		tx, err := bsvord.LoadTx(c.Param("txid"))
		if err != nil {
			handleError(c, err)
		}
		ins, err := bsvord.ProcessInsTx(tx, 0, 0)
		if err != nil {
			handleError(c, err)
		}
		c.JSON(http.StatusOK, ins)
	})

	r.GET("/api/inscriptions/:txid", func(c *gin.Context) {
		tx, err := bsvord.LoadTx(c.Param("txid"))
		if err != nil {
			handleError(c, err)
		}
		ins, err := bsvord.ProcessInsTx(tx, 0, 0)
		if err != nil {
			handleError(c, err)
		}
		c.Header("cache-control", "max-age=604800,immutable")
		c.JSON(http.StatusOK, ins)
	})

	r.GET("/api/files/origins/:origin", func(c *gin.Context) {
		origin, err := hex.DecodeString(c.Param("origin"))
		if err != nil {
			handleError(c, err)
			return
		}
		ins, err := bsvord.LoadInsByOrigin(origin)
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
	if httpErr, ok := err.(*bsvord.HttpError); ok {
		c.String(httpErr.StatusCode, "%v", httpErr.Err)
	} else {
		c.String(http.StatusInternalServerError, "%v", err)
	}
}
