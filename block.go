package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"fmt"
)

// TODO: For now there will be no blocks and transactions will be mined individually

type TxO struct {
	Value uint64
	To    []byte
}

type TxI struct {
	// TODO: Add signature

	From      []byte
	TxHash    [sha256.Size]byte
	OutputIdx uint32
}

type PoW struct {
	Nonce uint64
	Hash  [sha256.Size]byte
}

type Tx struct {
	Inputs  []TxI
	Outputs []TxO
	PoW
}

func NewTx(inputs []TxI, outputs []TxO) *Tx {
	return &Tx{
		Inputs:  inputs,
		Outputs: outputs,
	}
}

func (tx *Tx) Serialize() []byte {
	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(tx)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func TxDeserialize(raw []byte) *Tx {
	tx := Tx{}
	buf := bytes.Buffer{}
	buf.Write(raw)
	decoder := gob.NewDecoder(&buf)
	err := decoder.Decode(&tx)
	if err != nil {
		panic(err)
	}
	return &tx
}

func (tx *Tx) Print() {
	fmt.Println("TRANSACTION:")
	fmt.Println("\tInputs:")
	for _, in := range tx.Inputs {
		fmt.Printf("\t\tFrom: %x found in %x at offset %d\n", in.From, in.TxHash, in.OutputIdx)
	}
	fmt.Println("\tOutputs:")
	for _, out := range tx.Outputs {
		fmt.Printf("\t\t%d to %x\n", out.Value, out.To)
	}
	fmt.Println("\tProof of Work:")
	fmt.Printf("\t\tNonce: %d\n", tx.PoW.Nonce)
	fmt.Printf("\t\tPoW hash: %x\n", tx.PoW.Hash)
}

// Get the binary payload of the transaction for use in the PoW
func (tx *Tx) Payload() []byte {
	var parts [][]byte
	for _, in := range tx.Inputs {
		outIdxRaw := make([]byte, 4)
		binary.LittleEndian.PutUint32(outIdxRaw, in.OutputIdx)
		parts = append(parts, in.From, in.TxHash[:], outIdxRaw)
	}
	for _, out := range tx.Outputs {
		valRaw := make([]byte, 8)
		binary.LittleEndian.PutUint64(valRaw, out.Value)
		parts = append(parts, out.To, valRaw)
	}
	return bytes.Join(parts, []byte{})
}

func (tx *Tx) Mine() {
	fmt.Println("Mining block...")
	// The number of 0 bits the hash needs to start with
	var difficulty [sha256.Size]byte
	difficulty[2] = 0b00000100
	tx.PoW.Nonce = 0
	nonceRaw := make([]byte, 8)
	for {
		binary.LittleEndian.PutUint64(nonceRaw, tx.PoW.Nonce)
		sum := sha256.Sum256(append(tx.Payload(), nonceRaw...))
		if bytes.Compare(difficulty[:], sum[:]) > 0 {
			// Valid Nonce
			tx.PoW.Hash = sum
			break
		} else {
			// Invalid Nonce
			tx.PoW.Nonce += 1
		}
	}
	fmt.Println("Success!")
	fmt.Println("Nonce: " + fmt.Sprintf("%d", tx.PoW.Nonce))
	fmt.Println("PoW: " + fmt.Sprintf("%x", tx.PoW.Hash))
}
