Blockchain
==========

> A blockchain in Go. Written as a learning exercise.

The transaction model is based on the one used by bitcoin.
A transaction consumes a set of transaction outputs and produces (usually two) new outputs.

A proof of work algorithm with static difficulty is used to achieve distributed consensus.
There is a mining reward as incentive for running a node.

The [Ed25519](https://ed25519.cr.yp.to/) signature algorithm provides transaction authorization and SHA-256 is used for block chaining and the proof of work.

The [bbolt](https://pkg.go.dev/go.etcd.io/bbolt) key/value store is used as a persistence layer and stores

- The blockchain itself as a mapping from block PoW hash to block
- Unspent transaction outputs (UTxOs) as a mapping from public key hash to UTxOs belonging to the keypair.
  This is used as an optimization to avoid full chain traversal when determining account balance or creating transactions.
- A keystore as a mapping from public key hash to public key
- A reference to the latest block in the chain

There is currently no block limit and there are no transaction fees.

TODO
----

- **Networking** — Nodes do not yet communicate with one another.
- **User interface** — Probably an HTTP interface for clients to interact with the network using a node as a gateway

