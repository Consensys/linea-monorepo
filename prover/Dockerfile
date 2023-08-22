# syntax=docker/dockerfile:1
# Required to use the --build-context flag

###############################
#
# G O + R U S T   B U I L D E R
#
###############################
FROM golang:alpine as go-builder

RUN apk add --no-cache rust cargo bash make
WORKDIR /usr/src/
COPY --from=prover      . ./prover
COPY --from=constraints . ./constraints
COPY --from=corset      . ./corset
ENV CARGO_REGISTRIES_CRATES_IO_PROTOCOL sparse
RUN cd prover && make bin/prover && make bin/prover-large && make bin/checker


#######################
#
# F I N A L   I M A G E
#
#######################
FROM alpine:3

ARG VERSION
ENV PROVER_VERSION=${VERSION}

RUN apk add --no-cache bash musl-dev libgcc
RUN mkdir -p /opt/linea/prover /data/work  /app/logs

# Import the LICENSE
COPY ./LICENSE /opt/linea/prover/

# Import the prover
COPY --from=go-builder /usr/src/prover/bin/checker /opt/linea/prover/checker
COPY --from=go-builder /usr/src/prover/bin/prover /opt/linea/prover/prover
COPY --from=go-builder /usr/src/prover/bin/prover-large /opt/linea/prover/prover-large

# Import the proving key and the compiled R1CS
RUN mkdir -p /opt/linea/prover/setup/full
RUN mkdir -p /opt/linea/prover/setup/full-large

# Import the plumbing
COPY ./prover/docker/prover.sh  /opt/linea/prover/prover.sh
COPY ./prover/docker/prover-large.sh  /opt/linea/prover/prover-large.sh
COPY ./prover/docker/eternal.sh /opt/linea/prover/eternal.sh

# Add the environment variables (in light mode, this is ignored)
ENV PROVER_PKEY_FILE /opt/linea/prover/setup/full/proving-key
ENV PROVER_VKEY_FILE /opt/linea/prover/setup/full/verifying-key
ENV PROVER_R1CS_FILE /opt/linea/prover/setup/full/r1cs
ENV PROVER_SOL_VERIFIER /opt/linea/prover/setup/full/verifierContract.sol


# # More environment variable to be set from the deployment
#
# ## Directory where to look for the conflated traces
# # ENV PROVER_CONFLATED_TRACES_DIR /shared/prover/conflated-traces
#
# ## If sets, the prover will skip the verification of the traces and generate
# ## a dummy proof. This overrides the behavior of DEV_LIGHT_PROVER.
# # ENV PROVER_SKIP_TRACES true
#
# ##  CHAIN_ID of the layer 2
# # ENV LAYER2_CHAIN_ID 58190
#
# ## ADDRESS of the bridge on layer 2
# # ENV LAYER2_MESSAGE_SERVICE_CONTRACT 0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
#
# ## Verbosity of the logs
# # ENV LOG_LEVEL info
#
# ## Prover features (only used for the full prover)
ENV PROVER_WITH_STATE_MANAGER true
ENV PROVER_WITH_KECCAK true
# # ENV PROVER_WITH_ECDSA false

# In order to disable one of the indices, put it to -1
ENV PROVER_VERIFIER_INDEX_LIGHT 0
ENV PROVER_VERIFIER_INDEX_FULL 1
ENV PROVER_VERIFIER_INDEX_FULL_LARGE 2

## Set to true to activate the dev-light version
ENV PROVER_DEV_LIGHT_VERSION false
ENV LISTING_LIMIT 100

CMD ["bash", "-c", "/opt/linea/prover/eternal.sh $WORKER_ID $BINARY $VERSION $FROM_DIR $IF_EXT $TO_DIR $OF_EXT $DONE_DIR $LOGS_DIR /opt/linea/prover/prover-large.sh $LISTING_LIMIT $LARGE_EXT $RETRYABLE_ERROR_CODES"]
