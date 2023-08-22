#!/bin/bash
bin/testcase-gen \
    --chainid=1 \
    --seed=47 \
    --l2-bridge-address=0xc499a572640b64ea1c8c194c43bc3e19940719dc \
    --ofile=./5-8-zkProof.json \
    --start-from-block 5 \
    --start-from-timestamp=150000100 \
    --start-from-root-hash=0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb \
    --max-l1-l2-msg-receipt-per-block 16 \
    --max-l2-l1-logs-per-block 16 \
    --num-blocks 4 \
    --min-tx-byte-size 0 \
    --max-tx-byte-size 314 \
    --max-tx-per-block 10 \
    --end-with-root-hash=0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc
