package main

import (
	"crypto/sha256"
	"fmt"
)

func main() {
	fmt.Println("Starting")
	bc, err := NewBlockchain(sha256.Sum256([]byte("Alice")))
	if err != nil {
		panic(err)
	}
	defer bc.Close()

	block := bc.MineNext()

	retrievedBlock, _ := bc.GetBlock(block.PoW.Hash[:])
	retrievedBlock.Print()

	fmt.Println(bc.GetBalance(bc.user))
}
