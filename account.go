package main

import (
	"crypto/ed25519"
	"crypto/sha256"
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
