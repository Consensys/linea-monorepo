FROM golang:1.22-alpine as builder

RUN apk add --no-cache git && git clone https://github.com/protolambda/eth2-testnet-genesis.git \
    && cd eth2-testnet-genesis \
    && git checkout 4b3498476f14b872b43080eee319adea45286daf \
    && go install .

FROM alpine:3.19.1

RUN  apk add --no-cache bash

COPY --from=builder /go/bin/eth2-testnet-genesis /usr/local/bin/
