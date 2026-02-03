# General Rules

**Impact: LOW (improves code quality and maintainability)**

## Avoid Magic Numbers

Use named constants instead of hardcoded values:

```solidity
// Incorrect: magic numbers
function withdraw() external {
  require(block.timestamp > 1704067200, "Too early");
  require(amount <= 1000000000000000000, "Too much");
}

// Correct: named constants
uint256 public constant WITHDRAWAL_START = 1704067200;
uint256 public constant MAX_WITHDRAWAL = 1 ether;

function withdraw() external {
  require(block.timestamp > WITHDRAWAL_START, "Too early");
  require(amount <= MAX_WITHDRAWAL, "Too much");
}
```

## Assembly

Use hex for memory offsets:

```solidity
// Correct: hex offsets
assembly {
  mstore(add(mPtr, 0x20), _var)
  mstore(add(mPtr, 0x40), _otherVar)
}

// Incorrect: decimal offsets
assembly {
  mstore(add(mPtr, 32), _var)
}
```

## Linting

Run linting before committing:

```bash
pnpm run lint:fix
```

---

## Deployment

### Environment Variables

- Set `VERIFY_CONTRACT=true` for block explorer verification
- Use network-specific private keys (e.g., `SEPOLIA_PRIVATE_KEY`)
- See `contracts/docs/deployment.md` for full parameter reference

### Deployment Commands

```bash
# Deploy to testnet
pnpm -F contracts run deploy:sepolia

# Verify contract
pnpm -F contracts run verify:sepolia
```

---

## Testing

### Linting Commands

```bash
# Solidity linting
pnpm -F contracts run lint:sol

# TypeScript linting
pnpm -F contracts run lint:ts

# Fix all lint issues
pnpm run lint:fix
```

### Test Commands

```bash
# Run tests with coverage
pnpm -F contracts run coverage

# Run specific test file
pnpm -F contracts run test -- --grep "MessageService"
```

---

## Inheritance & Customization

When extending Linea contracts:

- Use `virtual`/`override` keywords
- Override `CONTRACT_VERSION()` for custom versions
- See examples in `src/_testing/unit/` for patterns

```solidity
// Base contract
function getVersion() public pure virtual returns (string memory) {
  return "1.0.0";
}

// Extended contract
function getVersion() public pure override returns (string memory) {
  return "1.1.0";
}
```

**Note**: Any modifications from audited code should be independently audited.
