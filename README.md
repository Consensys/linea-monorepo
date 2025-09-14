# Status Network

<a href="https://x.com/StatusL2">
  <img src="https://img.shields.io/badge/X-%23000000.svg?style=for-the-badge&logo=X&logoColor=white" alt="X Follow" height="20">
</a>
<a href="https://github.com/status-im/status-network-monorepo/blob/main/LICENSE-APACHE">
  <img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="Apache 2.0 License" height="20">
</a>
<a href="https://github.com/status-im/status-network-monorepo/blob/main/LICENSE-MIT">
  <img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="MIT License" height="20">
</a>

This is the principal Status Network repository. 
On top of the Linea stack, it adds the smart contracts and infrastructure for Status Network's **gasless transaction system** powered by **RLN (Rate Limiting Nullifier) technology of Vac**. 
The additional Status Network features are optional, configurable using the CLI options (details provided under [Configuration Options](#configuration-options)).
Open-sourced under the [Apache 2.0](LICENSE-APACHE) and the [MIT](LICENSE-MIT) licenses.

## Quickstart

### Prerequisites
- Docker and Docker Compose
- Java 21 (Temurin)
- Make

### Start local network
- Only Status Network contracts on Linea stack (no extra Linea contract deployments):

```bash
STATUS_NETWORK_CONTRACTS_ENABLED=true LINEA_PROTOCOL_CONTRACTS_ONLY=true make start-env-with-rln
```

- Full Linea + Status Network (includes Linea protocol suite like token bridges):

```bash
LINEA_PROTOCOL_CONTRACTS_ONLY=false STATUS_NETWORK_CONTRACTS_ENABLED=true make start-env-with-rln-and-contracts
```

When the stack comes up, contract addresses are printed and can be verified with scripts under `contracts/local-deployments-artifacts/`.

### Rebuild after sequencer changes
Any edits to the sequencer or validators require rebuilding and restarting:

```bash
./build-rln-enabled-sequencer.sh
make clean-environment
STATUS_NETWORK_CONTRACTS_ENABLED=true LINEA_PROTOCOL_CONTRACTS_ONLY=true make start-env-with-rln
```

## What is Status Network?

[Status Network](https://status.network) is the **first natively gasless Ethereum L2**, optimized for social apps and games, featuring sustainable public funding for developers through native yield and DEX fees. Built on the Linea zkEVM stack, it provides high-performance, gas-free transactions while ensuring economic sustainability through a novel funding model and spam prevention technology.

## Gasless Transaction System

Status Network introduces **gasless transactions** through a **RLN technology** and **Karma reputation system**. This allows users to submit transactions without paying gas fees while maintaining network security and preventing spam.

### How It Works

**Rate Limiting Nullifier**: A cryptographic system that prevents spam by limiting transaction rates through nullifier-based proofs. Implementation can be found [here](https://github.com/vacp2p/zerokit).

**火 Karma System**: A reputation-based mechanism where users earn Karma soulbound tokens through positive network participation. Users will have different levels of daily gasless transaction quota depending on their Karma amount. Contract code implementation can be found [here](https://github.com/vacp2p/staking-reward-streamer).

**Premium Gas Bypass**: When users exceed their daily gasless transaction quota, they can still submit transactions by paying premium gas fees. The premium gas threshold is configurable and allows users to bypass rate limiting restrictions when needed.

### Architecture Components

#### Besu Plugin Components
- [**TxForwarder Transaction Pool Validator**](besu-plugins/linea-sequencer/sequencer/src/main/java/net/consensys/linea/sequencer/txpoolvalidation/validators/RlnProverForwarderValidator.java): Forwards incoming transaction data to the [RLN prover service](https://github.com/vacp2p/status-rln-prover) to generate RLN proofs based on the user Karma
- [**Modified LineaEstimateGas RPC**](besu-plugins/linea-sequencer/sequencer/src/main/java/net/consensys/linea/rpc/methods/LineaEstimateGas.java): Dynamically provides zero gas or premium gas estimates based on real-time user karma and usage quotas

#### Sequencer Components
- [**RLNVerifier Transaction Pool Validator**](besu-plugins/linea-sequencer/sequencer/src/main/java/net/consensys/linea/sequencer/txpoolvalidation/validators/RlnVerifierValidator.java): Verifies incoming transactions from RPC nodes using RLN proofs received from the prover service
- [**RLN Bridge**](besu-plugins/linea-sequencer/sequencer/src/main/rust/rln_bridge/src/lib.rs): JNI interface for RLN proof verification, providing high-performance cryptographic verification of rate limiting nullifiers
- [**Nullifier Tracking**](besu-plugins/linea-sequencer/sequencer/src/main/java/net/consensys/linea/sequencer/txpoolvalidation/shared/NullifierTracker.java): High-performance tracking to prevent double-spending and nullifier reuse
- [**Deny List Management**](besu-plugins/linea-sequencer/sequencer/src/main/java/net/consensys/linea/sequencer/txpoolvalidation/shared/DenyListManager.java): Shared deny list manager providing single source of truth for deny list state

#### Configuration Options

The Status Network RLN validator system can be configured using various CLI options defined in [LineaRlnValidatorCliOptions.java](besu-plugins/linea-sequencer/sequencer/src/main/java/net/consensys/linea/config/LineaRlnValidatorCliOptions.java).

- **`--plugin-linea-rln-enabled`**: Enable/disable RLN validation for gasless transactions (default: `false`)
- **`--plugin-linea-rln-verifying-key`**: Path to the RLN verifying key file (required when RLN is enabled)
- **`--plugin-linea-rln-proof-service`**: RLN Proof service endpoint in `host:port` format (default: `localhost:50051`)
- **`--plugin-linea-rln-karma-service`**: Karma service endpoint in `host:port` format (default: `localhost:50052`)
- **`--plugin-linea-rln-deny-list-path`**: Path to the gasless deny list file (default: `/var/lib/besu/gasless-deny-list.txt`)

### Troubleshooting
- If a 0-gas (gasless) transaction is accepted by the RPC but not included, check the sequencer logs for RLN proof cache timeouts or epoch mismatches and ensure `--plugin-linea-rln-epoch-mode=TEST` is used in local demos.
- If paid transactions fail for “min gas price” or “upfront cost” locally, ensure L2 genesis has `baseFeePerGas=0` and sequencer `min-gas-price=0` in config.
- Premium gas transactions (gas > configured threshold) bypass RLN validation by design.

### CI
- GitHub Actions runs sequencer unit tests (Java 21) with Gradle caching. JNI-dependent RLN native tests are excluded from CI.

## How to contribute

Contributions are welcome!

### Guidelines for Non-Code and other Trivial Contributions
Please keep in mind that we do not accept non-code contributions like fixing comments, typos or some other trivial fixes. Although we appreciate the extra help, managing lots of these small contributions is unfeasible, and puts extra pressure in our continuous delivery systems (running all tests, etc). Feel free to open an issue pointing to any of those errors, and we will batch them into a single change.

1. [Create an issue](https://github.com/status-im/status-network-monorepo/issues)
> If the proposed update is non-trivial, also tag us for discussion.
2. Submit the update as a pull request from your [fork of this repo](https://github.com/status-im/status-network-monorepo/fork), and tag us for review.
> Include the issue number in the pull request description and (optionally) in the branch name.

Consider starting with a ["good first issue"](https://github.com/status-im/status-network-monorepo/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22).

Before contributing, ensure you're familiar with:

- Our [contribution guide](docs/contribute.md)
- Our [code of conduct](docs/code-of-conduct.md)
- Our [Security policy](docs/security.md)

### Useful links

- [Status Network home](https://status.network)
- [Status Network docs](https://docs.status.network)
- [Telegram builder's community](https://t.me/statusl2)
- [X](https://x.com/StatusL2)
