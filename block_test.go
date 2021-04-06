package main

import (
	"testing"

	"crypto/sha256"

	"github.com/google/go-cmp/cmp"
)

func createTx() *Tx {
	acc, _ := NewAccount()
	return NewTx(
		[]TxI{
			{
				From: acc.Id,
				Output: &TxOPath{
					BlockHash: sha256.Sum256([]byte("Some hash")),
					TxIdx:     1,
					OutputIdx: 5,
				},
			},
		},
		[]TxO{
			{
				Value: 123,
				To:    sha256.Sum256([]byte("Someone else")),
			},
		},
		map[AccountId]Signature{
			AccountId(emptyHash): []byte{0xFF},
		},
	)
}

func createBlock() *Block {
	return NewBlock()

}

func TestSerialization(t *testing.T) {
	org_block := &Block{
		Transactions: []*Tx{
			createTx(),
		},
		PoW: &PoW{
			Nonce: 123,
			Hash:  sha256.Sum256([]byte("Something")),
		},
	}
	ser := org_block.Serialize()
	des_block := BlockDeserialize(ser)

	if !cmp.Equal(org_block, des_block) {
		t.Log(cmp.Diff(org_block, des_block))
		t.Fail()
	}
}
