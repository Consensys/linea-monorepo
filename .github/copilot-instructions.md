# GitHub Copilot Instructions

Instructions for GitHub Copilot when working on the Linea monorepo.

## Documentation

This repository has centralized documentation in `/documentation/`.

For architectural questions, consult:
- System overview: `/documentation/architecture/OVERVIEW.md`
- Component details: `/documentation/components/{component}.md`
- Development guide: `/documentation/development/README.md`

## Repository Structure

| Path | Language | Purpose |
|------|----------|---------|
| `contracts/` | Solidity | L1/L2 smart contracts |
| `coordinator/` | Kotlin | Orchestration service |
| `prover/` | Go | ZK proof generation |
| `tracer/` | Java | EVM trace generation |
| `besu-plugins/` | Kotlin | Besu plugin extensions |
| `sdk/` | TypeScript | Developer SDK |
| `postman/` | TypeScript | Message relay service |
| `bridge-ui/` | Next.js | Bridge frontend |
| `e2e/` | TypeScript | End-to-end tests |
| `jvm-libs/` | Kotlin | Shared JVM libraries |
| `ts-libs/` | TypeScript | Shared TS utilities |
| `tracer-constraints/` | Lisp | ZK constraints |

## Key Terms

- **Shnarf**: Commitment hash for blob verification
- **Conflation**: Grouping L2 blocks into batches
- **Batch**: Group of blocks for proof generation
- **Blob**: EIP-4844 data container
- **Aggregation**: Combined ZK proof
- **Finalization**: L1 state commitment
- **Traces**: EVM execution matrices

## Code Style

### Solidity
- OpenZeppelin patterns for upgrades
- Events for all state changes
- NatSpec documentation

### Kotlin
- ktlint formatting
- Vert.x for async operations

### TypeScript
- ESLint + Prettier
- Strict TypeScript

### Go
- Standard Go formatting
- gnark for ZK circuits

## Common Patterns

### Contract Upgrades
Uses TransparentUpgradeableProxy pattern. See `/contracts/src/proxies/`.

### Cross-Chain Messaging
L1→L2: sendMessage → anchor → claim
L2→L1: sendMessage → finalize → claim with proof

### Prover Communication
File-based: Coordinator writes to `requests/`, Prover writes to `responses/`.
