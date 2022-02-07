package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

var (
	blockChain     = NewBlockChain()
	nodeIdentifier = strings.ReplaceAll(uuid.New().String(), "-", "")
)

func main() {
	r := NewRoute()

	_ = r.Run(":8777")
}

func NewRoute() *gin.Engine {
	r := gin.Default()

	r.GET("/chain", chain)
	r.POST("/transactions/new", newTransaction)
	r.POST("/mine", mine)

	return r
}

func chain(c *gin.Context) {
	c.JSON(http.StatusOK, formatResp(map[string]interface{}{
		"node":   nodeIdentifier,
		"chain":  blockChain,
		"length": len(blockChain.Blocks),
	}, 0, "ok"))
}

func newTransaction(c *gin.Context) {
	var transaction Transaction
	err := c.BindJSON(&transaction)
	if err != nil {
		c.JSON(http.StatusOK, formatResp("", 10000, fmt.Sprintf("unexpected input data: %v", err)))
		return
	}
	index := blockChain.AddTransaction(transaction.Sender, transaction.Recipient, transaction.Amount)
	c.JSON(http.StatusOK, formatResp(fmt.Sprintf("Transaction will be added to Block %d", index), 0, "ok"))
}

func mine(c *gin.Context) {
	proof := blockChain.ProofOfWork(blockChain.lastBlock().Proof)

	blockChain.AddTransaction("0", nodeIdentifier, 1)

	block := blockChain.NewBlock(proof, nil)

	c.JSON(http.StatusOK, formatResp(block, 0, "ok"))
}

func formatResp(data interface{}, code int64, msg string) map[string]interface{} {
	return map[string]interface{}{
		"data": data,
		"code": code,
		"msg":  msg,
	}
}
