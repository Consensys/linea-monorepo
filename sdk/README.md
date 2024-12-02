# Linea SDK

## Description

The Linea SDK package is a comprehensive solution consisting of two distinct parts that facilitate message delivery between Ethereum and Linea networks. It provides functionality for interacting with contracts and retrieving; message status and information.

## Part 1: SDK

The Linea SDK is the first component of the package. It focuses on interacting with smart contracts on both Ethereum and Linea networks and provides custom functions to obtain message information. Notable features of the Linea SDK include:

- Feature 1: Getting contract instances and addresses
- Feature 2: Getting message information by message hash
- Feature 3: Getting messages by transaction hash
- Feature 4: Getting a message status by message hash
- Feature 5: Claiming messages

## Part 2: Postman

The Postman component is the second part of the package and enables the delivery of messages between Ethereum and Linea networks. It offers the following key features:

- Feature 1: Listening for message sent events on Ethereum
- Feature 2: Listening for message hash anchoring events to check if a message is ready to be claimed
- Feature 3: Automatic claiming of messages with a configurable retry mechanism
- Feature 4: Checking receipt status for each transaction

All messages are stored in a configurable Postgres DB.

## Installation

To install this package, execute the following command:

`npm install @consensys/linea-sdk`

## Usage

This package exposes two main classes for usage:
- The `PostmanServiceClient` class is used to run a Postman service for delivering messages.
- The `LineaSDK` class is used to interact with smart contracts on Ethereum and Linea (both Sepolia and Mainnet).

## License

This package is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for more information.
