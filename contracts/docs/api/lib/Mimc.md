# Solidity API

## Mimc

### DataMissing

```solidity
error DataMissing()
```

Thrown when the data is not provided

### DataIsNotMod32

```solidity
error DataIsNotMod32()
```

Thrown when the data is not purely in 32 byte chunks

### FR_FIELD

```solidity
uint256 FR_FIELD
```

### hash

```solidity
function hash(bytes _msg) external pure returns (bytes32 mimcHash)
```

Performs a MiMC hash on the data provided

_Only data that has length modulus 32 is hashed, reverts otherwise_

#### Parameters

| Name | Type | Description |
| ---- | ---- | ----------- |
| _msg | bytes | The data to be hashed |

#### Return Values

| Name | Type | Description |
| ---- | ---- | ----------- |
| mimcHash | bytes32 | The computed MiMC hash |

