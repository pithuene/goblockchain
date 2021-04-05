package main

import (
	bolt "go.etcd.io/bbolt"
)

type UTxO struct {
	Value uint64
	Path  TxOPath
}

func (bc *Blockchain) GetUTxO() map[SHA256Sum][]*UTxO {
	// TODO: Use the utxo bucket to cache this
	uTxOs := make(map[SHA256Sum][]*UTxO)

	bc.db.View(func(t *bolt.Tx) error {
		bucket := t.Bucket([]byte(chainBucketName))

		removeOutputForInput := func(in TxI) {
			// The list of outputs by the sender
			sendersOutputs := uTxOs[in.From]

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
				uTxOs[in.From] = sendersOutputs[:len(sendersOutputs)-1]
			}
		}

		currBlockHash := bc.latestBlock
		var currBlock *Block
		for currBlockHash != nullHash {
			currBlock = BlockDeserialize(bucket.Get(bc.latestBlock[:]))

			for txHash, tx := range currBlock.Transactions {
				for _, in := range tx.Inputs {
					removeOutputForInput(in)
				}
				for outIdx, out := range tx.Outputs {
					uTxOs[out.To] = append(uTxOs[out.To], &UTxO{
						Value: out.Value,
						Path: TxOPath{
							BlockHash: currBlockHash,
							TxHash:    txHash,
							OutputIdx: uint32(outIdx),
						},
					})
				}
			}

			currBlockHash = currBlock.LastBlockHash
		}

		return nil
	})
	return uTxOs
}

func (bc *Blockchain) GetBalance(user SHA256Sum) uint64 {
	uTxOSet := bc.GetUTxO()
	uTxOs := uTxOSet[user]
	var balance uint64
	for _, uTxO := range uTxOs {
		balance += uTxO.Value
	}
	return balance
}
