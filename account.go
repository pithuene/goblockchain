package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/gob"
)

// Hash of the public key
type AccountId SHA256Sum

type Signature []byte

type Account struct {
	Id         AccountId
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

func NewAccount() (*Account, error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, err
	}
	id := sha256.Sum256(pub)
	return &Account{
		Id:         id,
		PublicKey:  pub,
		PrivateKey: priv,
	}, nil
}

func (acc *Account) Sign(tx *Tx) Signature {
	txHash := tx.Hash()
	return ed25519.Sign(acc.PrivateKey, txHash[:])
}

func (acc *Account) Serialize() []byte {
	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(acc)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func AccountDeserialize(raw []byte) *Account {
	var acc Account
	buf := bytes.Buffer{}
	buf.Write(raw)
	decoder := gob.NewDecoder(&buf)
	err := decoder.Decode(&acc)
	if err != nil {
		panic(err)
	}
	return &acc
}
