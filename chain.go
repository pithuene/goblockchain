package main

import (
	"crypto/sha256"
	"errors"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

type SHA256Sum [sha256.Size]byte

type Mempool struct {
	pool []*Tx
}

type Blockchain struct {
	db            *bolt.DB
	mempool       *Mempool
	latestBlock   SHA256Sum
	miningAccount *Account
}

const (
	dbFile             string = "blockchain.db"
	chainBucketName    string = "chain"
	utxoBucketName     string = "utxo"
	keystoreBucketName string = "keystore"
	miningReward       uint64 = 100
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
func NewBlockchain(miningAcc *Account) (*Blockchain, error) {
	db, err := bolt.Open(dbFile, 0666, nil)
	if err != nil {
		return nil, err
	}

	recreateBucket(db, chainBucketName)
	recreateBucket(db, utxoBucketName)
	recreateBucket(db, keystoreBucketName)

	mempool := NewMempool()

	bc := Blockchain{
		db:            db,
		mempool:       mempool,
		latestBlock:   nullHash,
		miningAccount: miningAcc,
	}

	bc.GenerateUTxO()

	fmt.Println("GENESIS")
	// Mine genesis block
	if _, err := bc.MineNext(); err != nil {
		return nil, err
	}

	bc.AddKey(miningAcc.PublicKey)

	return &bc, nil
}

func NewMempool() *Mempool {
	return &Mempool{
		pool: make([]*Tx, 0, 10),
	}
}

func (mp *Mempool) Push(tx *Tx) {
	mp.pool = append(mp.pool, tx)
}

func (mp *Mempool) Pop() (*Tx, error) {
	if len(mp.pool) == 0 {
		return nil, errors.New("Mempool empty")
	}
	tx := mp.pool[0]
	mp.pool = mp.pool[1:]
	return tx, nil
}

func (mp *Mempool) Count() int {
	return len(mp.pool)
}

// Creates a transaction for `value` coins and pushes it into the mempool
func (bc *Blockchain) Send(from *Account, to AccountId, value uint64) error {
	fmt.Printf("'%x' is sending %d to '%x'\n", from.Id, value, to)
	utxos := bc.GetUTxOsForUser(from.Id)
	sufficientFunds := utxos.Balance() >= value
	if sufficientFunds {
		var currValue uint64
		inputs := make([]TxI, 0)
		for _, utxo := range *utxos {
			if currValue >= value {
				break
			} else {
				inputs = append(inputs, TxI{
					From:   from.Id,
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
			To:    from.Id,
		}
		tx := NewTx(
			inputs,
			[]TxO{
				changeOutput,
				output,
			},
			make(map[AccountId]Signature),
		)
		sig := from.Sign(tx)
		tx.Signatures[from.Id] = sig
		bc.mempool.Push(tx)
		return nil
	} else {
		return errors.New(fmt.Sprintf("'%x' has insufficient funds to send %d coins", from.Id, value))
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

func (bc *Blockchain) MineNext() (*Block, error) {
	block := NewBlock()

	// Add mining reward transaction
	// The mining reward transaction is always the first transaction in a block
	miningRewardTx := NewTx(
		[]TxI{},
		[]TxO{
			{
				To:    bc.miningAccount.Id,
				Value: miningReward,
			},
		},
		map[AccountId]Signature{},
	)
	block.AddTransaction(miningRewardTx)

	// TODO: Later the decision which transactions to mine should be made based on fees
	for bc.mempool.Count() > 0 {
		tx, err := bc.mempool.Pop()
		if err != nil {
			return nil, err
		}
		if err := bc.VerifyTransaction(tx); err != nil {
			return nil, err
		}
		block.AddTransaction(tx)
	}

	block.LastBlockHash = bc.latestBlock
	block.Mine()

	bc.AddBlock(block)
	return block, nil
}

func (bc *Blockchain) GetBlock(powHash SHA256Sum) (*Block, error) {
	var block *Block
	err := bc.db.View(func(t *bolt.Tx) error {
		bucket := t.Bucket([]byte(chainBucketName))
		raw := bucket.Get(powHash[:])
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
