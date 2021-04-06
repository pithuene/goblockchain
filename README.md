Blockchain
==========

A blockchain in Go.

Architecture
------------

- `bbolt` as a persistence layer
    - Bucket for a mapping from block PoW hash to block
    - Bucket for unspent transaction outputs (UTxO)
    - Keystore bucket for AccountId to public key mapping
- UTxO and Bitcoin-like transactions
- Proof of Work system

