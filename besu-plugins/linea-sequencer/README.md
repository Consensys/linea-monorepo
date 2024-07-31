# Besu Plugins related to tracer and sequencer functionality

This repository hosts the implementation of the sequencer, the component of the Linea stack responsible for ordering transactions and building blocks, as well as executing them. It provides a set of [Hyperledger Besu](https://github.com/hyperledger/besu):Linea plugins. 

It serves developers by making the Linea tech stack open source under 
the [Apache 2.0 license](LICENSE).

## What is Linea?

[Linea](https://linea.build) is a developer-ready layer 2 network scaling Ethereum. It's secured with a zero-knowledge rollup, built on lattice-based cryptography, and powered by [Consensys](https://consensys.io).

## Get started

If you already have an understanding of the tech stack, use our [Quickstart](docs/quickstart.md) guide.

### Looking for Plugins?

Discover [existing plugins](docs/plugins.md) and understand the [plugin release process](docs/plugin-release.md). 

## Looking for the Linea code?

Linea's stack is made up of multiple repositories, these include:
- This repo, [linea-sequencer](https://github.com/Consensys/linea-sequencer): A set of Linea-Besu plugins for the sequencer and RPC nodes
- [linea-monorepo](https://github.com/Consensys/linea-monorepo): The main repository for the Linea stack & network 
- [linea-besu](https://github.com/Consensys/linea-besu): Fork of Besu to implement the Linea-Besu client
- [linea-tracer](https://github.com/Consensys/linea-tracer): Linea-Besu plugin which produces the traces that the constraint system applies and that serve as inputs to the prover
- [linea-constraints](https://github.com/Consensys/linea-constraints): Implementation of the constraint system from the specification
- [linea-specification](https://github.com/Consensys/linea-specification): Specification of the constraint system defining Linea's zk-EVM

Linea abstracts away the complexity of this technical architecture to allow developers to:

- [Bridge tokens](https://docs.linea.build/developers/guides/bridge)
- [Deploy a contract](https://docs.linea.build/developers/quickstart/deploy-smart-contract)
- [Run a node](https://docs.linea.build/developers/guides/run-a-node)

... and more.

## How to contribute

Contributions of any kind are welcome!

1. [Create an issue](https://github.com/Consensys/linea-sequencer/issues).
> If the proposed update is non-trivial, also tag us for discussion.
2. Submit the update as a pull request from your [fork of this repo](https://github.com/Consensys/linea-sequencer/fork), and tag us for review. 
> Include the issue number in the pull request description and (optionally) in the branch name.

Consider starting with a ["good first issue"](https://github.com/ConsenSys/linea-sequencer/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22).

Before contributing, ensure you're familiar with:

- Our [Linea contribution guide](https://github.com/Consensys/linea-monorepo/blob/main/docs/contribute.md)
- Our [Linea code of conduct](https://github.com/Consensys/linea-monorepo/blob/main/docs/code-of-conduct.md)
- The [Besu contribution guide](https://github.com/Consensys/linea-monorepo/blob/main/https://wiki.hyperledger.org/display/BESU/Coding+Conventions), for Besu:Linea related contributions
- Our [Security policy](https://github.com/Consensys/linea-monorepo/blob/main/docs/security.md)


### Useful links

- [Linea docs](https://docs.linea.build)
- [Linea blog](https://linea.mirror.xyz)
- [Support](https://support.linea.build)
- [Discord](https://discord.gg/linea)
- [Twitter](https://twitter.com/LineaBuild)
