package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Block struct {
	Index        int64          `json:"index"`
	Timestamp    time.Time      `json:"timestamp"`
	Transactions []*Transaction `json:"transactions"`
	Proof        int64          `json:"proof"`
	PreviousHash string         `json:"previous_hash"`
}

type Transaction struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Amount    int64  `json:"amount"`
}

type BlockChain struct {
	Blocks              []*Block       `json:"blocks"`
	CurrentTransactions []*Transaction `json:"current_transactions"`
	Nodes               map[string]struct{}
}

func NewBlockChain() *BlockChain {
	bc := &BlockChain{
		Blocks:              make([]*Block, 0),
		CurrentTransactions: make([]*Transaction, 0),
		Nodes:               make(map[string]struct{}),
	}
	previousHash := "0"
	bc.NewBlock(100, &previousHash)
	return bc
}

func (bc *BlockChain) NewBlock(proof int64, previousHash *string) *Block {
	prevHash := ""
	if previousHash == nil {
		prevHash = BlockHash(bc.lastBlock())
	} else {
		prevHash = *previousHash
	}
	block := &Block{
		Index:        bc.lastBlock().Index + 1,
		Timestamp:    time.Now(),
		Transactions: bc.CurrentTransactions,
		Proof:        proof,
		PreviousHash: prevHash,
	}
	bc.Blocks = append(bc.Blocks, block)
	bc.CurrentTransactions = make([]*Transaction, 0)
	return block
}

func (bc *BlockChain) AddTransaction(sender, recipient string, amount int64) int64 {
	bc.CurrentTransactions = append(bc.CurrentTransactions, &Transaction{
		Sender:    sender,
		Recipient: recipient,
		Amount:    amount,
	})
	return bc.lastBlock().Index + 1
}

func (bc *BlockChain) ProofOfWork(lastProof int64) int64 {
	proof := int64(0)
	for !bc.ValidProof(lastProof, proof) {
		proof += 1
	}
	return proof
}

func (bc *BlockChain) ValidProof(lastProof, proof int64) bool {
	proofStr := fmt.Sprintf("%d%d", lastProof, proof)
	proofHash := fmt.Sprintf("%x", sha256.Sum256([]byte(proofStr)))
	return proofHash[0:4] == "0000"
}

func (bc *BlockChain) RegisterNode(address string) {
	bc.Nodes[address] = struct{}{}
}

func (bc *BlockChain) TotalNode() []string {
	totalNode := make([]string, 0)
	for k := range bc.Nodes {
		totalNode = append(totalNode, k)
	}
	return totalNode
}

func (bc *BlockChain) ResolveConflicts() bool {
	maxLen := len(bc.Blocks)
	var newBlocks []*Block

	for node := range bc.Nodes {
		blocks := bc.resolveBlocks(node)
		if len(blocks) > maxLen && bc.validChain(blocks) {
			maxLen = len(blocks)
			newBlocks = blocks
		}
	}

	if len(newBlocks) > 0 {
		bc.Blocks = newBlocks
		return true
	}
	return false
}

func (bc *BlockChain) validChain(blocks []*Block) bool {
	lastBlock := blocks[0]
	for i := 1; i < len(blocks); i++ {
		if blocks[i].PreviousHash != BlockHash(lastBlock) {
			return false
		}
		if !bc.ValidProof(lastBlock.Proof, blocks[i].Proof) {
			return false
		}
		lastBlock = blocks[i]
	}
	return true
}

func (bc *BlockChain) resolveBlocks(node string) []*Block {
	resp, err := http.Get(fmt.Sprintf("%s/chain", node))
	if err != nil {
		log.Printf("Get %s chain error: %v", node, err)
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Get %s body error: %v", node, err)
		return nil
	}
	var chainSt struct {
		Data struct {
			Chain []*Block `json:"chain"`
		} `json:"data"`
	}
	err = json.Unmarshal(body, &chainSt)
	if err != nil {
		log.Printf("Get %s body unmarshal error: %v", node, err)
		return nil
	}
	return chainSt.Data.Chain
}

func (bc *BlockChain) lastBlock() *Block {
	if len(bc.Blocks) == 0 {
		return &Block{}
	}
	return bc.Blocks[len(bc.Blocks)-1]
}

func BlockHash(block *Block) string {
	blockByt, _ := json.Marshal(block)
	return fmt.Sprintf("%x", sha256.Sum256(blockByt))
}
