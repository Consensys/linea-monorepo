# Solidity API

## L1MessageManager

### rollingHashes

```solidity
mapping(uint256 => bytes32) rollingHashes
```

Contains the L1 to L2 messaging rolling hashes mapped to message number computed on L1.

### _messageClaimedBitMap

```solidity
struct BitMaps.BitMap _messageClaimedBitMap
```

This maps which message numbers have been claimed to prevent duplicate claiming.

### l2MerkleRootsDepths

```solidity
mapping(bytes32 => uint256) l2MerkleRootsDepths
```

Contains the L2 messages Merkle roots mapped to their tree depth.

### _addRollingHash

```solidity
function _addRollingHash(uint256 _messageNumber, bytes32 _messageHash) internal
```

Take an existing message hash, calculates the rolling hash and stores at the message number.

#### Parameters

| Name | Type | Description |
| ---- | ---- | ----------- |
| _messageNumber | uint256 | The current message number being sent. |
| _messageHash | bytes32 | The hash of the message being sent. |

### _setL2L1MessageToClaimed

```solidity
function _setL2L1MessageToClaimed(uint256 _messageNumber) internal
```

Set the L2->L1 message as claimed when a user claims a message on L1.

#### Parameters

| Name | Type | Description |
| ---- | ---- | ----------- |
| _messageNumber | uint256 | The message number on L2. |

### _addL2MerkleRoots

```solidity
function _addL2MerkleRoots(bytes32[] _newRoots, uint256 _treeDepth) internal
```

Add the L2 Merkle roots to the storage.

_This function is called during block finalization.
The _treeDepth does not need to be checked to be non-zero as it is,
already enforced to be non-zero in the circuit, and used in the proof's public input._

#### Parameters

| Name | Type | Description |
| ---- | ---- | ----------- |
| _newRoots | bytes32[] | New L2 Merkle roots. |
| _treeDepth | uint256 |  |

### _anchorL2MessagingBlocks

```solidity
function _anchorL2MessagingBlocks(bytes _l2MessagingBlocksOffsets, uint256 _currentL2BlockNumber) internal
```

Emit an event for each L2 block containing L2->L1 messages.

_This function is called during block finalization._

#### Parameters

| Name | Type | Description |
| ---- | ---- | ----------- |
| _l2MessagingBlocksOffsets | bytes | Is a sequence of uint16 values, where each value plus the last finalized L2 block number. indicates which L2 blocks have L2->L1 messages. |
| _currentL2BlockNumber | uint256 | Last L2 block number finalized on L1. |

### isMessageClaimed

```solidity
function isMessageClaimed(uint256 _messageNumber) external view returns (bool isClaimed)
```

Checks if the L2->L1 message is claimed or not.

#### Parameters

| Name | Type | Description |
| ---- | ---- | ----------- |
| _messageNumber | uint256 | The message number on L2. |

#### Return Values

| Name | Type | Description |
| ---- | ---- | ----------- |
| isClaimed | bool | Returns whether or not the message with _messageNumber has been claimed. |

