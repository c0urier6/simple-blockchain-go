package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

var (
	blockChain     = NewBlockChain()
	nodeIdentifier = strings.ReplaceAll(uuid.New().String(), "-", "")

	port = flag.String("p", "8777", "port")
)

func main() {
	flag.Parse()
	r := NewRoute()

	_ = r.Run(":" + *port)
}

func NewRoute() *gin.Engine {
	r := gin.Default()

	r.GET("/chain", chain)
	r.POST("/transactions/new", newTransaction)
	r.POST("/mine", mine)
	r.POST("/nodes/register", register)
	r.POST("/nodes/resolve", resolve)
	return r
}

func chain(c *gin.Context) {
	c.JSON(http.StatusOK, formatResp(map[string]interface{}{
		"chain":  blockChain.Blocks,
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

func register(c *gin.Context) {
	nodes := make([]string, 0)
	err := c.BindJSON(&nodes)
	if err != nil {
		c.JSON(http.StatusOK, formatResp("", 10000, fmt.Sprintf("unexpected input data: %v", err)))
		return
	}
	for _, node := range nodes {
		blockChain.RegisterNode(node)
	}
	c.JSON(http.StatusOK, formatResp(map[string]interface{}{
		"total_nodes": blockChain.TotalNode(),
	}, 0, "ok"))
}

func resolve(c *gin.Context) {
	replaced := blockChain.ResolveConflicts()
	c.JSON(http.StatusOK, formatResp(map[string]interface{}{
		"replaced": replaced,
		"chain":    blockChain.Blocks,
		"length":   len(blockChain.Blocks),
	}, 0, "ok"))
}

func formatResp(data interface{}, code int64, msg string) map[string]interface{} {
	return map[string]interface{}{
		"data": data,
		"code": code,
		"msg":  msg,
	}
}
