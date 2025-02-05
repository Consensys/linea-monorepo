FROM golang:1.23-alpine as builder

RUN apk add --no-cache git && git clone https://github.com/protolambda/eth2-testnet-genesis.git \
    && cd eth2-testnet-genesis \
    && go install .

FROM alpine:3.20.3

RUN apk add --no-cache bash

COPY --from=builder /go/bin/eth2-testnet-genesis /usr/local/bin/
COPY ./generate-genesis.sh /config/generate-genesis.sh

ENTRYPOINT [ "/bin/sh" , "/config/generate-genesis.sh" ]
