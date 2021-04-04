Blockchain
==========

A blockchain in Go.

Architecture
------------

- `bbolt` as a persistence layer
    - Bucket for a mapping from block PoW hash to block
    - TODO: Bucket for unspent transaction outputs (UTxO)
- UTxO and Bitcoin-like transactions
- Proof of Work system

