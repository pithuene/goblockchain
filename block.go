package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"fmt"
)

type PoW struct {
	Nonce uint64
	Hash  SHA256Sum
}

type Block struct {
	Transactions  map[SHA256Sum]*Tx
	LastBlockHash SHA256Sum
	PoW           *PoW
}

func NewBlock() *Block {
	return &Block{
		Transactions: make(map[SHA256Sum]*Tx),
		PoW:          &PoW{},
	}
}

func (block *Block) AddTransaction(hash SHA256Sum, tx *Tx) {
	block.Transactions[hash] = tx
}

func (block *Block) Serialize() []byte {
	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(block)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func BlockDeserialize(raw []byte) *Block {
	var block Block
	buf := bytes.Buffer{}
	buf.Write(raw)
	decoder := gob.NewDecoder(&buf)
	err := decoder.Decode(&block)
	if err != nil {
		panic(err)
	}
	return &block
}

// Get the binary representation of the block for hashing purposes
// Last block hash must be set before.
func (block *Block) Binary() []byte {
	if block.LastBlockHash == emptyHash {
		// Make sure LastBlockHash is set before hashing
		panic("Tried getting binary of block with last block hash")
	}
	var binaryBlock []byte
	for _, tx := range block.Transactions {
		binaryBlock = append(binaryBlock, tx.Binary()...)
	}
	binaryBlock = append(binaryBlock, block.LastBlockHash[:]...)
	return binaryBlock
}

func (block *Block) Mine() {
	fmt.Println("Mining block...")

	// The number of 0 bits the hash needs to start with
	var difficulty SHA256Sum
	difficulty[2] = 0b00000100

	binaryBlock := block.Binary()

	block.PoW.Nonce = 0
	nonceRaw := make([]byte, 8)

	for {
		binary.LittleEndian.PutUint64(nonceRaw, block.PoW.Nonce)
		sum := sha256.Sum256(append(binaryBlock, nonceRaw...))
		if bytes.Compare(difficulty[:], sum[:]) > 0 {
			// Valid Nonce
			block.PoW.Hash = sum
			break
		} else {
			// Invalid Nonce
			block.PoW.Nonce += 1
		}
	}
	fmt.Println("Success!")
	block.Print()
}

// Print the block to stdout for debugging
func (block *Block) Print() {
	fmt.Println("BLOCK:")
	fmt.Println("\tTransactions:")
	for _, tx := range block.Transactions {
		tx.Print("\t\t")
	}
	fmt.Println("\tProof of Work:")
	fmt.Printf("\t\tNonce: %d\n", block.PoW.Nonce)
	fmt.Printf("\t\tPoW hash: %x\n", block.PoW.Hash)
}
