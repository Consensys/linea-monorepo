# Message Service Test Scenarios


## Roles

`LIMIT_MANAGER` controls daily withdrawal limits

`CONTRACT_UPGRADER` controls contract limits

- TimelockController for upgrades `CONTRACT_UPGRADER`
- Multisig for upgrades `LIMIT_MANAGER`

## Tests

### Initialize tests
 - Access control is set for `CONTRACT_UPGRADER`
 - Access control is set for `LIMIT_MANAGER`
 - Can't initialise a second time
 - `pauseManagerAddress` not `zero address`
 - `pauseManagerAddress` is set

### Access control tests
- add account to role ?
- remove account from role?
- cannot remove self as admin ??
- cannot set `self` as `CONTRACT_UPGRADER`
- cannot set `zero address` as `CONTRACT_UPGRADER`
- cannot set `self` as `LIMIT_MANAGER`
- cannot set `zero address` as `LIMIT_MANAGER`

### Upgrades
- can initiate upgrade
  - **note**: this has to be the TimeLockController


### Withdrawal Limits
- Can set limits as `LIMIT_MANAGER`
- Cannot set limits as non-`LIMIT_MANAGER`

### Send Message tests
- `paused` contract `reverts`

 **If not paused:**
- hash exists reverts
- value checks
  - `_value` sent + `_fee` = `msg.value`
- `_destinationChainId` allowed ( is there an allowed network list )
- `_to` is not empty
- `_value` and `_calldata` are both empty (pointless transport) ??
- `_nonce` is default and hash does not exist
- `MessageSent` is emitted on success
  - indexed params are set for topics
- `_destinationChainId` not same as `block.chainId`
- cannot reenter on send with same hash
- `cannot send` a message with a value if daily `withdrawal limit is reached`
- `can send` a message with a value if daily `withdrawal limit is not reached`

- **CRITICAL HACK HERE IF DONE WRONG** `cannot set destination` when destination is message manager on other layer - reverts
- **CRITICAL HACK HERE IF DONE WRONG** `cannot set destination` when destination is message manager proxy on other layer - reverts

### Claim Message tests
- `paused` contract `reverts`

 **If not paused:**
- hash does not exist reverts
- hash exists and is pending succeeds
- hash exists and is delivered reverts
- `BridgeMessageClaimed` emitted on success
- `_destinationChainId` matches `block.chainId`
- Call permutation tests
  - error on ETH send only reverts
  - error on contract call only reverts
  - error on contract call with ETH send reverts
  - no error on ETH send only does not revert
  - no error on contract call only does not revert
  - no error on contract call with ETH send does not rever
  - `zero fee` does not try send ETH to fee recipient
  - error on ETH Fee send reverts
  - no error on ETH Fee send does not revert
- `cannot claim` a message with a value if daily `withdrawal limit is reached`
- `can claim` a message with a value if daily `withdrawal limit is not reached`
- delivery `fee` and sent `value` increments daily withdrawal amounts used
- `_feeRecipient` is `zero address`, the fee is sent to `msg.sender`
- `_feeRecipient` is `not zero address`, , the fee is sent to `_feeRecipient`

### Reentry
- `cannot` enter with `claimMessage` -> `claimMessage`
- `can` enter `claimMessage` -> `sendMessage`

### setting networks tests
- can add network
- network exists reverts
- not empty value