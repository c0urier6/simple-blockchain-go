package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
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
}

func NewBlockChain() *BlockChain {
	bc := &BlockChain{
		Blocks:              make([]*Block, 0),
		CurrentTransactions: make([]*Transaction, 0),
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
