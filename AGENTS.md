# Agent Instructions

> Universal instructions for AI coding agents (Codex, Aider, and other tools).

## Documentation Location

**All architecture and component documentation is centralized in `/documentation/`.**

Before exploring code for architectural questions, read:

1. `/documentation/README.md` - Navigation index
2. `/documentation/architecture/OVERVIEW.md` - System architecture and data flows
3. `/documentation/components/README.md` - Component index

## Quick Reference

### Components

| Component | Path | Language | Documentation |
|-----------|------|----------|---------------|
| Smart Contracts | `contracts/` | Solidity | `/documentation/components/contracts.md` |
| Coordinator | `coordinator/` | Kotlin | `/documentation/components/coordinator.md` |
| Prover | `prover/` | Go | `/documentation/components/prover.md` |
| Tracer | `tracer/` | Java | `/documentation/components/tracer.md` |
| SDK | `sdk/` | TypeScript | `/documentation/components/sdk.md` |
| Postman | `postman/` | TypeScript | `/documentation/components/postman.md` |
| Bridge UI | `bridge-ui/` | TypeScript | `/documentation/components/bridge-ui.md` |
| Besu Plugins | `besu-plugins/` | Kotlin | `/documentation/components/besu-plugins.md` |
| Constraints | `tracer-constraints/` | Lisp | `/documentation/components/tracer-constraints.md` |

### Key Diagrams

All diagrams are in Mermaid format in `/documentation/diagrams/`:

- `system-architecture.mmd` - Overall system
- `proving-pipeline.mmd` - Block → Proof → Finalization
- `l1-to-l2-message-flow.mmd` - L1→L2 messaging
- `l2-to-l1-message-flow.mmd` - L2→L1 messaging
- `component-dependency-graph.mmd` - Package dependencies

## Glossary

Essential terms used throughout the codebase:

- **Shnarf**: Partially computed public input that links state root hashes to L2 data commitments
- **Conflation**: Grouping L2 blocks into batches for proving
- **Batch**: Group of blocks ready for proof generation
- **Blob**: EIP-4844 data container for L1 data availability
- **Aggregation**: Recursive ZK proof combining multiple proofs
- **Finalization**: L1 verification of ZK proof and state commitment
- **Rolling hash**: Recursive hash for L2→L1 message ordering
- **Anchoring**: The mechanism by which messaging data is placed on a destination chain to validate claiming against
- **Traces**: EVM execution matrices (tracer output, prover input)

## External Dependencies

Some components are in separate repositories:

- **Maru**: Execution engine for sequencer (Engine API)
- **Linea Besu**: Modified Besu fork for L2
- **go-corset**: Constraint compiler for ZK circuits

See `/documentation/architecture/EXTERNAL-DEPENDENCIES.md` for details.

## Common Tasks

### Build & Test

```bash
# Full local stack
make start-env-with-tracing-v2

# Clean environment
make clean-environment

# Run E2E tests
cd e2e && pnpm run test:e2e:local
```

### Package-Specific

```bash
# Contracts
cd contracts && npx hardhat test

# Coordinator
./gradlew :coordinator:app:build

# Prover
cd prover && go test ./...

# TypeScript packages
pnpm run build && pnpm run test
```

## PR Review Guidance

When reviewing PRs, consider:

1. **Contract changes**: BI compatibility, storage layout and upgrade safety, event emissions, migration - prover/coordinator interactions
2. **Coordinator changes**: Impact on proving pipeline, gas pricing
3. **Cross-package**: Check `/documentation/architecture/OVERVIEW.md` for interactions
4. **Config changes**: Verify against `/documentation/operations/README.md`
