export CFG=config/config-sepolia-full.toml
export REQ_DIR=~/sepolia-testing-full/prover-aggregation/requests
export RESPONSE_DIR=~/sepolia-testing-full/prover-aggregation/responses
# export LOG_DIR=~/sepolia-testing-full/prover-aggregation/logs
export LOG_DIR=logs/prover-aggregation

make bin/prover

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

# /home/ubuntu/prover-sepolia-full/traces/conflated/4454961-4455038.conflated.v0.8.0-rc3.lt
# /home/ubuntu/sepolia-testing-full/traces/conflated/4454961-4455038.conflated.v0.8.0-rc3.lt
