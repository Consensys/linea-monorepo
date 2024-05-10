 # Message Manager Test Scenarios


## Roles
- `ORIGIN_BATCH_SETTER` coordinator setting the origin block hashes
- `MESSAGE_SERVICE` for outbox messages

## Tests
### Initialize tests
 - Empty address check for `ORIGIN_BATCH_SETTER`
 - Empty address check for `MESSAGE_SERVICE`
 - Can't initialise a second time
 - `pauserRoleAddress` not `zero address`
 - `pauserRoleAddress` is set

### Get claimedMessages message status (public mapping)
- can get message status for claimed hash
- returns empty value for non-existing hash

### Add origin block and message hashes
 - reverts `if paused`

 **If not paused:**
 - can add non-existing batch of hashes as `ORIGIN_BATCH_SETTER`
 - can't add batch of hashes where an existing duplicate exists as `ORIGIN_BATCH_SETTER` - reverts
 - can't add batch of hashes as non-`ORIGIN_BATCH_SETTER` - reverts
 - Any empty reference (sent data) reverts the batch
 - Empty list of hashes reverts
 - Block number 0 reverts

### Add Single Message Reference set tests for outbox
 - reverts `if paused`

 **If not paused:**
 - can add non-existing as `MESSAGE_SERVICE` 
 - can't add duplicate existing as `MESSAGE_SERVICE` 
 - can't add duplicate existing as non-`MESSAGE_SERVICE` 
 - reference (sent data) cannot be empty - reverts 