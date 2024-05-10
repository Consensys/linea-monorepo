// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.24;

import { IPlonkVerifier } from "../interfaces/l1/IPlonkVerifier.sol";

/// @dev Test verifier contract that returns true.
contract IntegrationTestTrueVerifier is IPlonkVerifier {
  function Verify(bytes calldata, uint256[] calldata) external pure returns (bool) {
    return true;
  }
}
