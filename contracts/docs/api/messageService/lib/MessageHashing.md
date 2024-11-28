# Solidity API

## MessageHashing

### _hashMessage

```solidity
function _hashMessage(address _from, address _to, uint256 _fee, uint256 _valueSent, uint256 _messageNumber, bytes _calldata) internal pure returns (bytes32 messageHash)
```

Hashes messages using assembly for efficiency.

_Adding 0xc0 is to indicate the calldata offset relative to the memory being added to.
If the calldata is not modulus 32, the extra bit needs to be added on at the end else the hash is wrong._

#### Parameters

| Name | Type | Description |
| ---- | ---- | ----------- |
| _from | address | The from address. |
| _to | address | The to address. |
| _fee | uint256 | The fee paid for delivery. |
| _valueSent | uint256 | The value to be sent when delivering. |
| _messageNumber | uint256 | The unique message number. |
| _calldata | bytes | The calldata to be passed to the destination address. |

