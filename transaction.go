package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

type TxO struct {
	Value uint64
	To    SHA256Sum
}

type TxOPath struct {
	BlockHash SHA256Sum
	TxHash    SHA256Sum
	OutputIdx uint32
}

type TxI struct {
	// TODO: Add signature

	From   SHA256Sum
	Output *TxOPath
}

type Tx struct {
	Inputs  []TxI
	Outputs []TxO
}

func NewTx(inputs []TxI, outputs []TxO) *Tx {
	return &Tx{
		Inputs:  inputs,
		Outputs: outputs,
	}
}

func (txop *TxOPath) Binary() []byte {
	outputIdxBin := make([]byte, 4)
	binary.LittleEndian.PutUint32(outputIdxBin, txop.OutputIdx)
	return append(txop.BlockHash[:], append(txop.TxHash[:], outputIdxBin...)...)
}

// Get the binary representation of the transaction for hashing purposes
func (tx *Tx) Binary() []byte {
	var parts [][]byte
	for _, in := range tx.Inputs {
		parts = append(parts, in.From[:], in.Output.Binary())
	}
	for _, out := range tx.Outputs {
		valRaw := make([]byte, 8)
		binary.LittleEndian.PutUint64(valRaw, out.Value)
		parts = append(parts, out.To[:], valRaw)
	}
	return bytes.Join(parts, []byte{})
}

// Calculates the SHA-256 checksum of the transaction
func (tx *Tx) Hash() SHA256Sum {
	return sha256.Sum256(tx.Binary())
}

// Print the transaction to stdout for debugging
func (tx *Tx) Print(prefix string) {
	fmt.Printf("%sTRANSACTION %x\n", prefix, tx.Hash())
	fmt.Println(prefix + "\tInputs:")
	for _, in := range tx.Inputs {
		fmt.Println(prefix + "\t\tINPUT:")
		fmt.Printf("%s\t\t\tFrom: %x\n", prefix, in.From)
		fmt.Printf("%s\t\t\tOutput: %d\n of transaction %x in block %x", prefix, in.Output.OutputIdx, in.Output.TxHash, in.Output.BlockHash)
		fmt.Printf("%s\t\t\tTransaction: %x\n", prefix, in.Output.TxHash)
		fmt.Printf("%s\t\t\tBlock: %x\n", prefix, in.Output.BlockHash)
	}
	fmt.Println(prefix + "\tOutputs:")
	for _, out := range tx.Outputs {
		fmt.Printf("%s\t\t%d to %x\n", prefix, out.Value, out.To)
	}
}
