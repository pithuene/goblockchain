package main

import (
	"errors"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

type Mempool struct {
	pool []*Tx
}

type Blockchain struct {
	db      *bolt.DB
	mempool *Mempool
	user    []byte
}

const (
	dbFile          string = "blockchain.db"
	chainBucketName string = "chain"
	mempoolCapacity int    = 128
)

// Creates a blank blockchain.
// Overrides any existing chain.
// Creates the genesis block.
func NewBlockchain(user []byte) (*Blockchain, error) {
	db, err := bolt.Open(dbFile, 0666, nil)
	if err != nil {
		return nil, err
	}

	// Initialize blank bucket
	err = db.Update(func(tx *bolt.Tx) error {
		// Clear chain bucket
		if tx.Bucket([]byte(chainBucketName)) != nil {
			err := tx.DeleteBucket([]byte(chainBucketName))
			if err != nil {
				return err
			}
		}
		// Create blank bucket
		if _, err = tx.CreateBucket([]byte(chainBucketName)); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	mempool := NewMempool()

	// TODO: Remove all value here. For now this gives the creator 100 coins.
	// TODO: Coins should only be introduced as mining rewards, but as long as there are no blocks that is difficult.
	genesis := NewTx([]TxI{}, []TxO{{
		To:    user,
		Value: 100,
	}})

	mempool.Push(genesis)

	return &Blockchain{
		db,
		mempool,
		user,
	}, nil

	/*db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("Chain"))
		if err != nil {
			return fmt.Errorf("Create bucket: %s", err)
		}
		bucket.Put([]byte("key"), []byte("value"))
		return nil
	})*/
	/*
		db.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("Chain"))
			val := bucket.Get([]byte("key"))
			fmt.Printf("The value was: %s\n", val)
			return nil
		})*/
}

func NewMempool() *Mempool {
	return &Mempool{
		pool: make([]*Tx, mempoolCapacity),
	}
}

func (mp *Mempool) Push(tx *Tx) {
	mp.pool = append(mp.pool, tx)
}

func (mp *Mempool) Pop() *Tx {
	tx, pool := mp.pool[len(mp.pool)-1], mp.pool[:len(mp.pool)-1]
	mp.pool = pool
	return tx
}

func (bc *Blockchain) MineNext() *Tx {
	// TODO: Later the decision which transactions to mine should be made based on fees
	block := bc.mempool.Pop()
	block.Mine()

	// Add the mined block to chain
	bc.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(chainBucketName))
		bucket.Put(block.PoW.Hash[:], block.Serialize())
		return nil
	})
	return block
}

func (bc *Blockchain) GetBlock(powHash []byte) (*Tx, error) {
	var tx *Tx
	err := bc.db.View(func(t *bolt.Tx) error {
		bucket := t.Bucket([]byte(chainBucketName))
		raw := bucket.Get(powHash)
		if raw == nil {
			return errors.New("Block '" + fmt.Sprintf("%x", powHash) + "' not found!")
		}
		tx = TxDeserialize(raw)
		return nil
	})
	return tx, err
}

func (bc *Blockchain) Close() {
	bc.db.Close()
}
