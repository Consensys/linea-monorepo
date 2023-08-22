#!/bin/bash
bin/testcase-gen\
    --chainid=1 \
    --seed=74 \
    --l2-bridge-address=0xc499a572640b64ea1c8c194c43bc3e19940719dc \
    --ofile=./1-4-zkProof.json \
    --start-from-block 1 \
    --start-from-timestamp=150000000 \
    --start-from-root-hash=0x0af1a2b824be146c558a3b0d8a5f04a80e9a097ff85ac1a1b9fbe7c860de87f8 \
    --max-l1-l2-msg-receipt-per-block 16 \
    --max-l2-l1-logs-per-block 16 \
    --num-blocks 4 \
    --min-tx-byte-size 0 \
    --max-tx-byte-size 314 \
    --max-tx-per-block 10 \
    --end-with-root-hash=0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb

