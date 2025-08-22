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

This is the Status Network client repository. On top of the Linea stack, it adds the smart contracts and infrastructure for Status Network's **gasless transaction system** powered by **RLN (Rate Limiting Nullifier) technology of Vac**.
Open-sourced under the [Apache 2.0](LICENSE-APACHE) and the [MIT](LICENSE-MIT) licenses.

## What is Status Network?

[Status Network](https://status.network) is the **first natively gasless Ethereum L2**, optimized for social apps and games, featuring sustainable public funding for developers through native yield and DEX fees. Built on the Linea zkEVM stack, it provides high-performance, gas-free transactions while ensuring economic sustainability through a novel funding model and spam prevention technology.

## Gasless Transaction System

Status Network introduces **gasless transactions** through a **RLN (Rate Limiting Nullifier) technology** and **Karma reputation system**. This allows users to submit transactions without paying gas fees while maintaining network security and preventing spam.

### How It Works

**RLN (Rate Limiting Nullifier)**: A cryptographic system that prevents spam by limiting transaction rates through nullifier-based proofs. Implementation can be found [here](https://github.com/vacp2p/status-rln-prover).

**火 Karma System**: A reputation-based mechanism where users earn Karma soulbound tokens through positive network participation. Users will have different levels of daily gasless transaction quota depending on their Karma amount. Contract code implementation can be found [here](https://github.com/vacp2p/staking-reward-streamer).

### Key Features

- **🆓 Zero Gas Transactions**: Users with sufficient Karma balance can submit transactions without paying gas
- **🛡️ Rate Limiting**: RLN-based rate limiting prevents spam and abuse through cryptographic proofs
- **💰 Premium Gas Bypass**: Users who exceed their quota can bypass restrictions with premium gas payments
- **📊 Reputation-Based Access**: Karma tokens determine user privileges and transaction quotas

### Architecture Components

#### Prover Components
- [**TxForwarder Transaction Pool Validator**](besu-plugins/linea-sequencer/sequencer/src/main/java/net/consensys/linea/sequencer/txpoolvalidation/validators/RlnProverForwarderValidator.java): Forwards incoming transaction data to the RLN prover service to generate RLN proofs
- [**Enhanced LineaEstimateGas RPC**](besu-plugins/linea-sequencer/sequencer/src/main/java/net/consensys/linea/rpc/methods/LineaEstimateGas.java): Dynamically provides zero gas or premium gas estimates based on real-time user karma and usage quotas
- [**Karma Service Integration**](besu-plugins/linea-sequencer/sequencer/src/main/java/net/consensys/linea/sequencer/txpoolvalidation/shared/KarmaServiceClient.java): Real-time user tier and quota checking via gRPC

#### Verifier (Sequencer) Components
- [**RLNVerifier Transaction Pool Validator**](besu-plugins/linea-sequencer/sequencer/src/main/java/net/consensys/linea/sequencer/txpoolvalidation/validators/RlnVerifierValidator.java): Verifies incoming transactions from RPC nodes using RLN proofs received from the prover service
- [**Nullifier Tracking**](besu-plugins/linea-sequencer/sequencer/src/main/java/net/consensys/linea/sequencer/txpoolvalidation/shared/NullifierTracker.java): High-performance tracking to prevent double-spending and nullifier reuse
- [**Deny List Management**](besu-plugins/linea-sequencer/sequencer/src/main/java/net/consensys/linea/sequencer/txpoolvalidation/shared/DenyListManager.java): Shared deny list manager providing single source of truth for deny list state

For detailed configuration options, see the [Status Network Configuration CLI Options](besu-plugins/linea-sequencer/sequencer/src/main/java/net/consensys/linea/config/LineaRlnValidatorCliOptions.java).

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
- [Status blog](https://status.app/blog)
- [X](https://x.com/StatusL2)
