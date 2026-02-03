# Commands

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