package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
)

type TxO struct {
	Value uint64
	To    AccountId
}

type TxOPath struct {
	BlockHash SHA256Sum
	TxIdx     uint32
	OutputIdx uint32
}

type TxI struct {
	From   AccountId
	Output *TxOPath
}

type Tx struct {
	Inputs  []TxI
	Outputs []TxO
	// Every party contributing an input signs a hash of the transaction
	Signatures map[AccountId]Signature
}

func NewTx(inputs []TxI, outputs []TxO, sigs map[AccountId]Signature) *Tx {
	return &Tx{
		Inputs:     inputs,
		Outputs:    outputs,
		Signatures: sigs,
	}
}

// Verifies that all inputs have a valid signature and,
// that value out equals value in.
// If the returned error is nil, the transaction is valid
func (bc *Blockchain) VerifyTransaction(tx *Tx) error {
	// Find all unique payers involved and get their public keys
	payers := make(map[AccountId]ed25519.PublicKey)
	for _, in := range tx.Inputs {
		if payers[in.From] == nil {
			pubKey, err := bc.GetKey(in.From)
			if err != nil {
				return err
			}
			payers[in.From] = pubKey
		}
	}

	// Verify that every payer has signed the transaction
	txHash := tx.Hash()
	for accId, pubKey := range payers {
		sig, signed := tx.Signatures[accId]
		if !signed {
			return errors.New(fmt.Sprintf("Transaction invalid! No signature by '%x'.\n", accId))
		}
		if !ed25519.Verify(pubKey, txHash[:], sig[:]) {
			return errors.New(fmt.Sprintf("Transaction invalid! Signature by '%x' is incorrect!\n", accId))
		}
	}

	// Verify that the transaction doesn't output more coins than the inputs provide
	var valOut uint64
	for _, out := range tx.Outputs {
		valOut += out.Value
	}
	var valIn uint64
	for _, in := range tx.Inputs {
		block, err := bc.GetBlock(in.Output.BlockHash)
		if err != nil {
			return errors.New(fmt.Sprintf("Unable to find block '%x'\n", in.Output.BlockHash))
		}

		if uint32(len(block.Transactions)) <= in.Output.TxIdx {
			return errors.New(fmt.Sprintf("Unable to get Tx %d in block '%x'\n", in.Output.TxIdx, in.Output.BlockHash))
		}
		refTx := block.Transactions[in.Output.TxIdx]

		if uint32(len(refTx.Outputs)) <= in.Output.OutputIdx {
			return errors.New(fmt.Sprintf("Unable to get output %d in transaction %d\n", in.Output.OutputIdx, in.Output.TxIdx))
		}
		out := refTx.Outputs[in.Output.OutputIdx]
		valIn += out.Value
	}
	if valOut != valIn {
		return errors.New(fmt.Sprintf("Transaction invalid! Value in (%d) does not match value out (%d)", valIn, valOut))
	}

	return nil
}

func (txop *TxOPath) Binary() []byte {
	outputIdxBin := make([]byte, 4)
	binary.LittleEndian.PutUint32(outputIdxBin, txop.OutputIdx)

	txIdxBin := make([]byte, 4)
	binary.LittleEndian.PutUint32(txIdxBin, txop.TxIdx)

	return append(txop.BlockHash[:], append(txIdxBin, outputIdxBin...)...)
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
	fmt.Printf("%sTRANSACTION\n", prefix)
	fmt.Println(prefix + "\tInputs:")
	for _, in := range tx.Inputs {
		fmt.Println(prefix + "\t\tINPUT:")
		fmt.Printf("%s\t\t\tFrom: %x\n", prefix, in.From)
		fmt.Printf("%s\t\t\tOutput: %d\n", prefix, in.Output.OutputIdx)
		fmt.Printf("%s\t\t\tTransaction: %d\n", prefix, in.Output.TxIdx)
		fmt.Printf("%s\t\t\tBlock: %x\n", prefix, in.Output.BlockHash)
	}
	fmt.Println(prefix + "\tOutputs:")
	for _, out := range tx.Outputs {
		fmt.Printf("%s\t\t%d to %x\n", prefix, out.Value, out.To)
	}
}
