FROM golang:1.23-alpine as builder

RUN apk add --no-cache git && git clone -b kaustinen-genesis --single-branch https://github.com/gballet/eth2-testnet-genesis.git \
    && cd eth2-testnet-genesis \
    && go install .

COPY ./generate-genesis.sh /generate-genesis.sh

FROM alpine:3.20.3

RUN  apk add --no-cache bash

COPY --from=builder /generate-genesis.sh /config/generate-genesis.sh
COPY --from=builder /go/bin/eth2-testnet-genesis /usr/local/bin/

ENTRYPOINT [ "/bin/sh" , "/config/generate-genesis.sh" ]
