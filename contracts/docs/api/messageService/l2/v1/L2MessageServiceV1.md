# Solidity API

## L2MessageServiceV1

### MINIMUM_FEE_SETTER_ROLE

```solidity
bytes32 MINIMUM_FEE_SETTER_ROLE
```

The role required to set the minimum DDOS fee.

### _messageSender

```solidity
address _messageSender
```

_The temporary message sender set when claiming a message._

### nextMessageNumber

```solidity
uint256 nextMessageNumber
```

### minimumFeeInWei

```solidity
uint256 minimumFeeInWei
```

### REFUND_OVERHEAD_IN_GAS

```solidity
uint256 REFUND_OVERHEAD_IN_GAS
```

### DEFAULT_SENDER_ADDRESS

```solidity
address DEFAULT_SENDER_ADDRESS
```

_The default message sender address reset after claiming a message._

### constructor

```solidity
constructor() internal
```

### sendMessage

```solidity
function sendMessage(address _to, uint256 _fee, bytes _calldata) external payable
```

Adds a message for sending cross-chain and emits a relevant event.

_The message number is preset and only incremented at the end if successful for the next caller._

#### Parameters

| Name | Type | Description |
| ---- | ---- | ----------- |
| _to | address | The address the message is intended for. |
| _fee | uint256 | The fee being paid for the message delivery. |
| _calldata | bytes | The calldata to pass to the recipient. |

### claimMessage

```solidity
function claimMessage(address _from, address _to, uint256 _fee, uint256 _value, address payable _feeRecipient, bytes _calldata, uint256 _nonce) external
```

Claims and delivers a cross-chain message.

__feeRecipient Can be set to address(0) to receive as msg.sender.
messageSender Is set temporarily when claiming and reset post._

#### Parameters

| Name | Type | Description |
| ---- | ---- | ----------- |
| _from | address | The address of the original sender. |
| _to | address | The address the message is intended for. |
| _fee | uint256 | The fee being paid for the message delivery. |
| _value | uint256 | The value to be transferred to the destination address. |
| _feeRecipient | address payable | The recipient for the fee. |
| _calldata | bytes | The calldata to pass to the recipient. |
| _nonce | uint256 | The unique auto generated message number used when sending the message. |

### setMinimumFee

```solidity
function setMinimumFee(uint256 _feeInWei) external
```

The Fee Manager sets a minimum fee to address DOS protection.

_MINIMUM_FEE_SETTER_ROLE is required to set the minimum fee._

#### Parameters

| Name | Type | Description |
| ---- | ---- | ----------- |
| _feeInWei | uint256 | New minimum fee in Wei. |

### sender

```solidity
function sender() external view returns (address originalSender)
```

_The _messageSender address is set temporarily when claiming._

#### Return Values

| Name | Type | Description |
| ---- | ---- | ----------- |
| originalSender | address | The original sender stored temporarily at the _messageSender address in storage. |

### distributeFees

```solidity
modifier distributeFees(uint256 _feeInWei, address _to, bytes _calldata, address _feeRecipient)
```

The unspent fee is refunded if applicable.

#### Parameters

| Name | Type | Description |
| ---- | ---- | ----------- |
| _feeInWei | uint256 | The fee paid for delivery in Wei. |
| _to | address | The recipient of the message and gas refund. |
| _calldata | bytes | The calldata of the message. |
| _feeRecipient | address |  |

