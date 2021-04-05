package main

import (
	"crypto/sha256"
	"errors"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

type SHA256Sum [sha256.Size]byte

type Mempool struct {
	pool map[SHA256Sum]*Tx
}

type Blockchain struct {
	db          *bolt.DB
	mempool     *Mempool
	latestBlock SHA256Sum
	user        SHA256Sum
}

const (
	dbFile          string = "blockchain.db"
	chainBucketName string = "chain"
	utxoBucketName  string = "utxo"
)

// An unset, all zero hash used for comparisons
var emptyHash SHA256Sum

// An explicitly null hash. Used for indicating there is no previous block in the genesis block.
var nullHash SHA256Sum = SHA256Sum{
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
}

// Removes and recreates the bucket with a given name
func recreateBucket(db *bolt.DB, bucketName string) error {
	// Initialize blank bucket
	err := db.Update(func(tx *bolt.Tx) error {
		// Clear chain bucket
		if tx.Bucket([]byte(bucketName)) != nil {
			if err := tx.DeleteBucket([]byte(bucketName)); err != nil {
				return err
			}
		}
		// Create blank bucket
		if _, err := tx.CreateBucket([]byte(bucketName)); err != nil {
			return err
		}
		return nil
	})
	return err
}

// Creates a blank blockchain.
// Overrides any existing chain.
// Creates the genesis block.
func NewBlockchain(user SHA256Sum) (*Blockchain, error) {
	db, err := bolt.Open(dbFile, 0666, nil)
	if err != nil {
		return nil, err
	}

	recreateBucket(db, chainBucketName)
	recreateBucket(db, utxoBucketName)

	mempool := NewMempool()

	// TODO: Remove all value here. For now this gives the creator 100 coins.
	// TODO: Coins should only be introduced as mining rewards, but as long as there are no blocks that is difficult.
	genesis := NewTx(
		[]TxI{},
		[]TxO{
			{
				To:    user,
				Value: 100,
			},
		},
	)

	mempool.Push(genesis)

	return &Blockchain{
		db:          db,
		mempool:     mempool,
		latestBlock: nullHash,
		user:        user,
	}, nil
}

func NewMempool() *Mempool {
	return &Mempool{
		pool: make(map[SHA256Sum]*Tx),
	}
}

func (mp *Mempool) Push(tx *Tx) {
	mp.pool[tx.Hash()] = tx
}

func (mp *Mempool) Pop() (SHA256Sum, *Tx) {
	for hash, tx := range mp.pool {
		delete(mp.pool, hash)
		return hash, tx
	}
	return emptyHash, nil
}

func (mp *Mempool) Count() int {
	return len(mp.pool)
}

// Creates a transaction for `value` coins and pushes it into the mempool
func (bc *Blockchain) Send(from SHA256Sum, to SHA256Sum, value uint64) error {
	utxos := bc.GetUTxOsForUser(from)
	sufficientFunds := utxos.Balance() >= value
	if sufficientFunds {
		var currValue uint64
		inputs := make([]TxI, 0)
		for _, utxo := range *utxos {
			if currValue >= value {
				break
			} else {
				inputs = append(inputs, TxI{
					From:   from,
					Output: &utxo.Path,
				})
				currValue += utxo.Value
			}
		}

		output := TxO{
			Value: value,
			To:    to,
		}

		change := currValue - value
		changeOutput := TxO{
			Value: change,
			To:    from,
		}
		tx := NewTx(
			inputs,
			[]TxO{
				changeOutput,
				output,
			},
		)
		bc.mempool.Push(tx)
		return nil
	} else {
		return errors.New(fmt.Sprintf("'%x' has insufficient funds to send %d coins", from, value))
	}
}

// Appends a block to the blockchain
// Fails if the blocks LastBlockHash doesn't match the latest block
// or if the block has not been mined yet
func (bc *Blockchain) AddBlock(block *Block) error {
	if block.LastBlockHash != bc.latestBlock {
		return errors.New("Block is not a valid extension of the chain")
	}
	if block.PoW.Hash == emptyHash {
		return errors.New("Block is not mined yet")
	}
	err := bc.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(chainBucketName))
		bucket.Put(block.PoW.Hash[:], block.Serialize())
		bc.latestBlock = block.PoW.Hash
		return nil
	})

	bc.UpdateUTxOSet(block)

	return err
}

func (bc *Blockchain) MineNext() *Block {
	block := NewBlock()
	// TODO: Later the decision which transactions to mine should be made based on fees
	for bc.mempool.Count() > 0 {
		hash, tx := bc.mempool.Pop()
		block.AddTransaction(hash, tx)
	}
	block.LastBlockHash = bc.latestBlock
	block.Mine()

	bc.AddBlock(block)
	return block
}

func (bc *Blockchain) GetBlock(powHash []byte) (*Block, error) {
	var block *Block
	err := bc.db.View(func(t *bolt.Tx) error {
		bucket := t.Bucket([]byte(chainBucketName))
		raw := bucket.Get(powHash)
		if raw == nil {
			return errors.New("Block '" + fmt.Sprintf("%x", powHash) + "' not found!")
		}
		block = BlockDeserialize(raw)
		return nil
	})
	return block, err
}

func (bc *Blockchain) Close() {
	bc.db.Close()
}
