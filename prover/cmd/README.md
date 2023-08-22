## Run the tools

### Generate the verifier contract

```
cd ./generate-contract
go run main.go .
```

### Generate the test-cases

```
make bin/testcase-gen
```

#### Usage

```
  -chainid int
        required, integer, the chainID of L2 (default -9223372036854775808)
  -end-with-root-hash string
        optional, root hash at which the sequence of blocks should end
  -l2-bridge-address string
        l2-bridge-address: required, the 0x prefixed address of the bridge
  -max-l1-l2-msg-receipt-per-block int
        optional, int, maximal number of l1 msg received per block (at most one batch per block). (default 16)
  -max-l2-l1-logs-per-block int
        optional, int, maximal number of l2 msg emitted per block (default 16)
  -max-tx-byte-size int
        optional, int, maximal size of an rlp encoded tx (default 512)
  -max-tx-per-block int
        optional, integer, maximal number of tx per block (default 20)
  -min-tx-byte-size int
        optional, int, minimal size of an rlp encoded tx (default 128)
  -num-blocks int
        optional, integer, maximal number of block in the conflated batch (default 10)
  -ofile string
        required, path, file where to write the testcase file
  -seed int
        required, integer, seed to use for the RNG (default -9223372036854775808)
  -start-from-block int
        required, integer (default -1)
  -start-from-root-hash string
        root upon which we generate a proof
  -start-from-timestamp uint
        required, UNIX-timestamp, timestamp starting which we generate the blocks
  -zeroes-in-calldata
        optional, bool, default to true. If set to true 75pc of the bytes of the calldata of the generated txs is set to zero (default true)

```