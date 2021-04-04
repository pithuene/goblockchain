package main

import (
	"testing"

	"bytes"
)

func TestSerialization(t *testing.T) {
	org_tx := NewTx(
		[]TxI{
			{},
		},
		[]TxO{},
	)
	t.Log(org_tx)
	ser := org_tx.Serialize()
	des_tx := TxDeserialize(ser)
	t.Log(des_tx)
	des_ser := des_tx.Serialize()
	res := bytes.Compare(ser, des_ser)

	if res != 0 {
		t.Fail()
	}
}
