# Sponsored EIP-7702 Script Design

## Goal

Add a sponsored EIP-7702 Type 4 transaction script and refactor shared logic out of
the self-sponsored script into `contracts/scripts/utils.ts`.

## Shared Functions (added to `contracts/scripts/utils.ts`)

- `requireEnv(name)` - reads env var or throws with clear message
- `checkDelegation(provider, address)` - checks `0xef0100` prefix, returns `{ isDelegated, implementationAddress? }`
- `getAccountInfo(provider, address)` - returns `{ address, balance, nonce }`
- `estimateGasFees(provider, rpcUrl, from, to?, data?)` - branches on Linea chain ID; uses `LineaEstimateGasClient` for Linea, `get1559Fees` otherwise
- `createAuthorization(signer, provider, targetContract, nonceOffset)` - signs authorization tuple; `nonceOffset` is +1 for self-sponsored (sender nonce incremented before auth processing), +0 for sponsored (separate authority)

## Self-Sponsored Script Changes

`contracts/scripts/testEIP7702/sendSelfSponsoredType4Tx.ts` - refactor to import and use
shared functions. Same behavior, less code.

## New Sponsored Script

`contracts/scripts/testEIP7702/sendSponsoredType4Tx.ts`

### Env Vars

| Variable | Required | Description |
|---|---|---|
| `RPC_URL` | yes | RPC endpoint |
| `SPONSOR_PRIVATE_KEY` | yes | Private key of the account that sends and pays for the tx |
| `AUTHORITY_PRIVATE_KEY` | yes | Private key of the account that signs the authorization |
| `TARGET_ADDRESS` | yes | Contract address to delegate to (set in authorization) |
| `TO_ADDRESS` | no | Transaction recipient; defaults to authority address |
| `CALLDATA` | no | Hex-encoded calldata; defaults to `0x` |

### Flow

1. Load env vars
2. Create sponsor wallet and authority wallet from their respective keys
3. Print account info for both
4. Check delegation status of authority address
5. Authority signs authorization for `TARGET_ADDRESS` (nonce offset +0)
6. Sponsor sends Type 4 tx with authorization list, to `TO_ADDRESS`, with `CALLDATA`
7. Wait for receipt, print result
