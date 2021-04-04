package main

import (
	"fmt"
)

func main() {
	fmt.Println("Starting")
	bc, err := NewBlockchain([]byte("Alice"))
	if err != nil {
		panic(err)
	}
	defer bc.Close()

	block := bc.MineNext()

	retrievedBlock, _ := bc.GetBlock(block.PoW.Hash[:])

	retrievedBlock.Print()
}
