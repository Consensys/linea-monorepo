// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.26;

import { IPlonkVerifier } from "../../verifiers/interfaces/IPlonkVerifier.sol";

/// @dev Test verifier contract that returns true.
contract TestPublicInputVerifier is IPlonkVerifier {
  uint256 public expectedPublicInput;

  constructor(uint256 _expectedPublicInput) {
    expectedPublicInput = _expectedPublicInput;
  }

  function Verify(bytes calldata, uint256[] calldata _publicInput) external view returns (bool) {
    return expectedPublicInput == _publicInput[0];
  }
}
