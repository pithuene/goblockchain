package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

type UTxO struct {
	Value uint64
	Path  TxOPath
}

// A slice of unspent transaction outputs
type UTxOs []*UTxO

type UTxOMap struct {
	Map        map[SHA256Sum]*UTxOs
	utxoBucket *bolt.Bucket
}

func NewUTxOMap(t *bolt.Tx) *UTxOMap {
	if !t.Writable() {
		panic("Can't create UTxO map from readonly transaction")
	}
	utxoBucket := t.Bucket([]byte(utxoBucketName))
	if utxoBucket == nil {
		panic("UTxO bucket not found!")
	}
	return &UTxOMap{
		Map:        make(map[SHA256Sum]*UTxOs),
		utxoBucket: utxoBucket,
	}
}

// Reads a user from the db into the map, overriding the current map value
// Returns an error if the user can not be found
func (utxoMap *UTxOMap) readUserFromDB(user SHA256Sum) error {
	raw := utxoMap.utxoBucket.Get(user[:])
	if raw == nil {
		return errors.New(fmt.Sprintf("User '%x' not found in UTxO bucket", user))
	}
	utxos := UTxOsDeserialize(raw)
	utxoMap.Map[user] = utxos
	return nil
}

func (utxoMap *UTxOMap) Get(user SHA256Sum) *UTxOs {
	// Check for cache hit
	utxos := utxoMap.Map[user]
	if utxos != nil {
		return utxos
	} else {
		if err := utxoMap.readUserFromDB(user); err != nil {
			// User is new (not present in memory or in db)
			utxoMap.Map[user] = &UTxOs{}
		}
		return utxoMap.Map[user]
	}
}

func (utxoMap *UTxOMap) Set(user SHA256Sum, utxos UTxOs) {
	utxoMap.Map[user] = &utxos
}

// Removes the outputs spent by a transaction from the map
func (utxoMap *UTxOMap) RemoveOutputsForInputs(tx *Tx) {
	for _, in := range tx.Inputs {
		// The list of outputs by the sender
		sendersOutputs := *utxoMap.Get(in.From)

		// Find index of the transaction output with matching path
		removeOutputIdx := -1
		for outIdx, output := range sendersOutputs {
			if output.Path.BlockHash == in.Output.BlockHash &&
				output.Path.TxHash == in.Output.TxHash &&
				output.Path.OutputIdx == in.Output.OutputIdx {
				removeOutputIdx = outIdx
				break
			}
		}

		// If there is no matching output the transaction would be invalid!
		if removeOutputIdx == -1 {
			panic("Did not find matching output for input in UTxO generation")
		} else {
			// Remove the spent output
			sendersOutputs[removeOutputIdx] = sendersOutputs[len(sendersOutputs)-1]
			utxoMap.Set(in.From, sendersOutputs[:len(sendersOutputs)-1])
		}
	}
}

func (utxoMap *UTxOMap) AddOutputs(tx *Tx, txHash SHA256Sum, blockHash SHA256Sum) {
	for outIdx, out := range tx.Outputs {
		utxos := append(*utxoMap.Get(out.To), &UTxO{
			Value: out.Value,
			Path: TxOPath{
				BlockHash: blockHash,
				TxHash:    txHash,
				OutputIdx: uint32(outIdx),
			},
		})
		utxoMap.Set(out.To, utxos)
	}
}

// Writes all values of the map back to the database
func (utxoMap *UTxOMap) Persist() {
	for owner, UTxOs := range utxoMap.Map {
		utxoMap.utxoBucket.Put(owner[:], UTxOs.Serialize())
	}
}

// (Re)creates the UTxO-Set by iterating over the entire blockchain
func (bc *Blockchain) GenerateUTxO() {
	// Empty the UTxO bucket
	if err := recreateBucket(bc.db, utxoBucketName); err != nil {
		panic(err)
	}

	bc.db.Update(func(t *bolt.Tx) error {
		chainBucket := t.Bucket([]byte(chainBucketName))
		utxoMap := NewUTxOMap(t)

		currBlockHash := bc.latestBlock
		var currBlock *Block
		for currBlockHash != nullHash {
			currBlock = BlockDeserialize(chainBucket.Get(currBlockHash[:]))

			for txHash, tx := range currBlock.Transactions {
				utxoMap.RemoveOutputsForInputs(tx)
				utxoMap.AddOutputs(tx, txHash, currBlockHash)
			}

			currBlockHash = currBlock.LastBlockHash
		}

		utxoMap.Persist()

		return nil
	})
}

// Changes the UTxO-Set by applying the transactions in a given block to it
func (bc *Blockchain) UpdateUTxOSet(block *Block) {
	bc.db.Update(func(t *bolt.Tx) error {
		utxoMap := NewUTxOMap(t)
		for txHash, tx := range block.Transactions {
			utxoMap.RemoveOutputsForInputs(tx)
			utxoMap.AddOutputs(tx, txHash, block.PoW.Hash)
		}
		return nil
	})
}

func (utxos UTxOs) Serialize() []byte {
	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(utxos)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func UTxOsDeserialize(rawUTxOs []byte) *UTxOs {
	var utxos UTxOs
	buf := bytes.Buffer{}
	buf.Write(rawUTxOs)
	decoder := gob.NewDecoder(&buf)
	err := decoder.Decode(&utxos)
	if err != nil {
		panic(err)
	}
	return &utxos
}

func (bc *Blockchain) getUTxOsForUser(user SHA256Sum) *UTxOs {
	var rawUTxOs []byte
	bc.db.View(func(t *bolt.Tx) error {
		utxoBucket := t.Bucket([]byte(utxoBucketName))
		rawUTxOs = utxoBucket.Get(user[:])
		return nil
	})
	return UTxOsDeserialize(rawUTxOs)
}

func (bc *Blockchain) GetBalance(user SHA256Sum) uint64 {
	uTxOs := bc.getUTxOsForUser(user)
	var balance uint64
	for _, uTxO := range *uTxOs {
		balance += uTxO.Value
	}
	return balance
}
