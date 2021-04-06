package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"errors"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

func (bc *Blockchain) AddKey(publicKey ed25519.PublicKey) error {
	return bc.db.Update(func(t *bolt.Tx) error {
		keystoreBucket := t.Bucket([]byte(keystoreBucketName))
		accountId := sha256.Sum256(publicKey)
		keystoreBucket.Put(accountId[:], publicKey)
		return nil
	})
}

func (bc *Blockchain) GetKey(accountId AccountId) (ed25519.PublicKey, error) {
	var pubKey ed25519.PublicKey
	err := bc.db.View(func(t *bolt.Tx) error {
		keystoreBucket := t.Bucket([]byte(keystoreBucketName))
		pubKey = keystoreBucket.Get(accountId[:])
		if pubKey == nil {
			return errors.New(fmt.Sprintf("Public key for %x not found in keystore\n", accountId))
		}
		return nil
	})
	return pubKey, err
}
