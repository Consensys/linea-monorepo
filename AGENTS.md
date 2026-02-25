# Linea Monorepo

Main repository for the Linea zkEVM stack: smart contracts for core protocol functions, ZK prover, coordinator for orchestration, Postman for bridge message execution, and Besu plugins for sequencer/RPC nodes.

Linea is a layer 2 network scaling Ethereum, secured with a zero-knowledge rollup built on lattice-based cryptography.

## Conventions

- Package manager: `pnpm` (not `npm`).
- Shared TypeScript utilities belong in `ts-libs/linea-shared-utils`. If a function is duplicated across projects, suggest moving it there.

## Self-Improvement

When you significantly course-correct during a task - wrong architecture, misunderstood pattern, missed convention - suggest adding the lesson to the appropriate AGENTS.md. Be specific: state what you got wrong, why, and what the rule should say. Scope project-specific gotchas to `<project>/AGENTS.md`.

## Project Guidelines

- [contracts/AGENTS.md](contracts/AGENTS.md) - Solidity contracts, deployment artifacts
- [e2e/AGENTS.md](e2e/AGENTS.md) - E2E tests, ABI generation
