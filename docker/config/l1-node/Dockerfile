FROM alpine:3.21

# Need "libc6-compat" for running glibc executable in alpine 
# Need "coreutils" for alpine date command to support -d option
RUN  apk add --no-cache bash \
    && apk add --no-cache libc6-compat \
    && apk add --no-cache coreutils

COPY ./generate-genesis.sh /config/generate-genesis.sh
COPY --from=ethpandaops/ethereum-genesis-generator:pk910-bash-el-genesis-generator /usr/local/bin/eth2-testnet-genesis /usr/local/bin/

ENTRYPOINT [ "/bin/bash", "/config/generate-genesis.sh" ]
