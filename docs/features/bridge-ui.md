# Bridge UI

> Next.js web application for bridging tokens between Ethereum and Linea.

## Overview

The Bridge UI is a production web application that provides a user-facing interface for ERC20 token bridging between Ethereum L1 and Linea L2. Built on Next.js 16 with React 19, it uses wagmi for wallet connection and the `@consensys/linea-sdk-viem` SDK for on-chain interactions.

## Stack

| Technology | Version | Role |
|------------|---------|------|
| Next.js | 16.1.5 | Framework |
| React | 19.1.5 | UI library |
| wagmi | 3.4.1 | Wallet connection and contract interaction |
| viem | catalog | Ethereum client |
| `@consensys/linea-sdk-viem` | workspace | Linea SDK integration |
| Playwright | — | E2E testing framework |
| Synpress | — | Web3 wallet automation for E2E |

## Features

- L1→L2 token deposits with approval flow
- L2→L1 token withdrawals with manual claiming
- Wallet connection (MetaMask, WalletConnect, etc.)
- Transaction status tracking
- Message claiming UI

## Project Structure

```
bridge-ui/
├── src/          # Application source
├── public/       # Static assets
├── test/
│   └── e2e/      # Playwright + Synpress E2E tests
├── Dockerfile    # Production container
├── next.config.ts
├── playwright.config.ts
└── package.json
```

## Test Coverage

| Test File | Runner | Validates |
|-----------|--------|-----------|
| `bridge-ui/test/e2e/bridge-l1-l2.spec.ts` | Playwright/Synpress | L1→L2 bridge flow in browser |
| `bridge-ui/test/e2e/bridge-l2-l1.spec.ts` | Playwright/Synpress | L2→L1 bridge flow in browser |
| `bridge-ui/test/e2e/general.spec.ts` | Playwright/Synpress | General UI behavior |

## Deployment

Containerized via `bridge-ui/Dockerfile`. Built and published via GitHub Actions workflow.
