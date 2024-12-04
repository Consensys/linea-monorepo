#!/bin/bash
# vim: set sw=4 sts=4 et foldmethod=indent :


REQUEST_PARAMS=$(cat << EOF
{
   "jsonrpc":"2.0",
   "method":"linea_getBlockTracesCountersV2",
   "params":[
      {
         "blockNumber":5000544,
         "expectedTracesEngineVersion":"0.2.0-rc4"
      }
   ],
   "id":1
}
EOF
)

curl --location --request POST 'http://localhost:8545' --data-raw "$REQUEST_PARAMS"
