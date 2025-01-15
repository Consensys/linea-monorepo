# Solidity API

## CallForwardingProxy

### target

```solidity
address target
```

The underlying target address that is called.

### constructor

```solidity
constructor(address _target) public
```

### fallback

```solidity
fallback() external payable
```

Defaults to, and forwards all calls to the target address.

