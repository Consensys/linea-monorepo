#!/bin/bash

TESTDIR=~/prover-testdata/sepolia-v0.8.0-rc8.1
CFG=config/config-sepolia-full.toml
JOB_LIST=("execution" "compression" "aggregation")

make bin/prover

# make setup &> run.setup.log

for job in "${JOB_LIST[@]}"; do

    REQ_DIR=$TESTDIR/prover-${job}/requests
    RESPONSE_DIR=$TESTDIR/prover-${job}/responses
    LOG_DIR=logs/prover-${job}

    rm -rf $RESPONSE_DIR # Clean the response folder
    rm -rf $LOG_DIR
    mkdir -p $LOG_DIR
    mkdir -p $RESPONSE_DIR

    export GOMEMLIMIT=700GiB

    for f in $(ls $REQ_DIR); do
        echo "[$(date)] handling request file $f"
        bin/prover prove --config $CFG --in $REQ_DIR/$f --out $RESPONSE_DIR/$f &> $LOG_DIR/$f.log
        EXIT_CODE=$?
        echo "[$(date)] finished with code $EXIT_CODE"
    done 

done