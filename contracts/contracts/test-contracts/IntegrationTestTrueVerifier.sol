// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.19;

import { IPlonkVerifier } from "../interfaces/IPlonkVerifier.sol";

/// @dev Test verifier contract that returns true.
contract IntegrationTestTrueVerifier is IPlonkVerifier {
  function Verify(bytes calldata, uint256[] calldata) external pure returns (bool) {
    return true;
  }
}
