package main

import (
	"testing"

	"crypto/sha256"

	"github.com/google/go-cmp/cmp"
)

func TestSerialization(t *testing.T) {
	tx := NewTx(
		[]TxI{
			{
				From: sha256.Sum256([]byte("Someone")),
				Output: &TxOPath{
					BlockHash: sha256.Sum256([]byte("Some hash")),
					TxHash:    sha256.Sum256([]byte("Some other hash")),
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
	)
	org_block := &Block{
		Transactions: map[SHA256Sum]*Tx{
			tx.Hash(): tx,
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
