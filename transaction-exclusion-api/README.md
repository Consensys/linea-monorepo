# Transaction Exclusion API Service
This micro-service will receive the transactions rejected by sequencer or other Linea Besu nodes,
persist them into a local database and expose a JSON-RPC v2 API to allow Linea users to query
why their transactions were not included.

## V1 API Methods
### linea_saveRejectedTransactionV1
```bash
curl -H 'content-type:application/json' --data '{
    "id": "1",
    "jsonrpc": "2.0",
    "method": "linea_saveRejectedTransactionV1",
    "params": {
        "txRejectionStage": "SEQUENCER",
        "timestamp": "2024-08-22T09:18:51Z",
        "blockNumber": 12345,
        "transactionRLP": "0x02f8388204d2648203e88203e88203e8941195cf65f83b3a5768f3c496d3a05ad6412c64b38203e88c666d93e9cc5f73748162cea9c0017b8201c8",
        "reasonMessage": "Transaction line count for module ADD=402 is above the limit 70",
        "overflows": [
            {
                "module": "ADD",
                "count": 402,
                "limit": 70
            },
            {
                "module": "MUL",
                "count": 587,
                "limit": 400
            }
        ]
    }
}' http://127.0.0.1:8082
```


### linea_getTransactionExclusionStatusV1
```bash
curl -H 'content-type:application/json' --data '{
    "jsonrpc": "2.0",
    "id": "53",
    "method": "linea_getTransactionExclusionStatusV1",
    "params": [
        "0xf5bf951edfefbaa6d9ed78c88942147cf98c8ef1f3d3416f99d2534675096569"
    ]
}' http://127.0.0.1:8082
```


### linea_saveRejectedTransactionV1 Response Examples:
The rejected transaction was successfully saved:
```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "status": "SAVED",
        "txHash": "0x526e56101cf39c1e717cef9cedf6fdddb42684711abda35bae51136dbb350ad7"
    }
}
```
The rejected transaction with the same tx hash and reason message was saved before:
```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "status": "DUPLICATE_ALREADY_SAVED_BEFORE",
        "txHash": "0x526e56101cf39c1e717cef9cedf6fdddb42684711abda35bae51136dbb350ad7"
    }
}
```
Invalid request params errors:
```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "error": {
        "code": -32602,
        "message": "Missing [timestamp] from the given request params"
    }
}
```
```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "error": {
        "code": -32602,
        "message": "Timestamp is not in ISO-8601: Text '2024-09-08T09:23:56Zdd' could not be parsed, unparsed text found at index 20"
    }
}
```
Other error:
```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "error": {
        "code": -32000,
        "message": "Server error",
        "data": "Database connection refused: localhost/127.0.0.1:5432"
    }
}
```

### linea_getTransactionExclusionStatusV1 Response Examples:
The rejected transaction was successfully found:
```json
{
    "jsonrpc": "2.0",
    "id": 53,
    "result": {
        "txHash": "0x526e56101cf39c1e717cef9cedf6fdddb42684711abda35bae51136dbb350ad7",
        "from": "0x4d144d7b9c96b26361d6ac74dd1d8267edca4fc2",
        "nonce": "0x64",
        "txRejectionStage": "SEQUENCER",
        "reasonMessage": "Transaction line count for module ADD=402 is above the limit 70",
        "blockNumber": "0x3039",
        "timestamp": "2024-08-22T09:18:51Z"
    }
}
```
The rejected transaction was not found:
```json
{
    "jsonrpc": "2.0",
    "id": 53,
    "result": null
}
```
Invalid request params errors:
```json
{
    "jsonrpc": "2.0",
    "id": 53,
    "error": {
        "code": -32602,
        "message": "Hex string of transaction hash cannot be parsed:  expected to have 32 bytes, but got 33"
    }
}
```
```json
{
    "jsonrpc": "2.0",
    "id": 53,
    "error": {
        "code": -32602,
        "message": "Hex string of transaction hash cannot be parsed: For input string: \"tt\" under radix 16"
    }
}
```
Other error:
```json
{
    "jsonrpc": "2.0",
    "id": 53,
    "error": {
        "code": -32000,
        "message": "Server error",
        "data": "Database connection refused: localhost/127.0.0.1:5432"
    }
}
```
