package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
)

type PoW struct {
	Nonce uint64
	Hash  SHA256Sum
}

var difficulty SHA256Sum = SHA256Sum{
	0x00, 0x00, 0b00000100, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

type Block struct {
	Transactions  []*Tx
	LastBlockHash SHA256Sum
	PoW           *PoW
}

func NewBlock() *Block {
	return &Block{
		Transactions: make([]*Tx, 0),
		PoW: &PoW{
			Nonce: 0,
			Hash:  emptyHash,
		},
	}
}

func (block *Block) AddTransaction(tx *Tx) {
	block.Transactions = append(block.Transactions, tx)
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

// Verifies the PoW aswell as all transactions
func (bc *Blockchain) VerifyBlock(block *Block) error {
	// Verify the PoW
	nonceRaw := make([]byte, 8)
	binary.LittleEndian.PutUint64(nonceRaw, block.PoW.Nonce)
	blockHash := sha256.Sum256(append(block.Binary(), nonceRaw...))
	if bytes.Compare(difficulty[:], blockHash[:]) <= 0 {
		// Invalid PoW
		return errors.New(fmt.Sprintf("Block invalid! The PoW is not valid."))
	}

	// Verify all transactions
	for txIdx, tx := range block.Transactions {
		if txIdx == 0 {
			// Mining reward transaction. This may mint new coins.
			if len(tx.Outputs) != 1 {
				return errors.New(fmt.Sprintf("Block invalid! Mining reward transaction has wrong number of outputs."))
			}
			if tx.Outputs[0].Value != miningReward {
				return errors.New(fmt.Sprintf("Block invalid! Mining reward transaction outputs invalid reward size."))
			}
		} else {
			if err := bc.VerifyTransaction(tx); err != nil {
				return err
			}
		}
	}

	return nil
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
