# Commands

## Linting

Run linting before committing:

```bash
pnpm run lint:fix
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