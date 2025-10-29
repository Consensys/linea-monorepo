// Copied verbatim from Lido audited contracts - https://github.com/lidofinance/core/blob/7cae7a14192ff094fb0eb089433ac9f6fd70e3c6/contracts/common/lib/BeaconTypes.sol

// SPDX-FileCopyrightText: 2025 Lido <info@lido.fi>
// SPDX-License-Identifier: GPL-3.0

// See contracts/COMPILERS.md
// solhint-disable-next-line lido/fixed-compiler-version
pragma solidity ^0.8.25;

struct Validator {
  bytes pubkey;
  bytes32 withdrawalCredentials;
  uint64 effectiveBalance;
  bool slashed;
  uint64 activationEligibilityEpoch;
  uint64 activationEpoch;
  uint64 exitEpoch;
  uint64 withdrawableEpoch;
}
struct BeaconBlockHeader {
  uint64 slot;
  uint64 proposerIndex;
  bytes32 parentRoot;
  bytes32 stateRoot;
  bytes32 bodyRoot;
}
