#!/bin/bash
# vim: set sw=4 sts=4 et foldmethod=indent :
REQUEST_PARAMS=$(cat << EOF
{
   "jsonrpc":"2.0",
   "method":"linea_generateConflatedTracesToFileV2",
   "params":[
      {
         "startBlockNumber":1043400,
         "endBlockNumber":1043455,
         "expectedTracesEngineVersion":"0.6.0-rc6"
      }
   ],
   "id":1
}
EOF
)

curl --location --request POST 'http://localhost:8545' --data-raw "$REQUEST_PARAMS"
