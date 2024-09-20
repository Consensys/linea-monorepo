// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.24;

import { IPlonkVerifier } from "../interfaces/l1/IPlonkVerifier.sol";

/// @dev Test verifier contract that returns true.
contract RevertingVerifier is IPlonkVerifier {
  enum Scenario {
    EMPTY_REVERT,
    GAS_GUZZLE
  }

  Scenario private scenario;

  constructor(Scenario _scenario) {
    scenario = _scenario;
  }

  function Verify(bytes calldata, uint256[] calldata) external returns (bool) {
    if (scenario == Scenario.GAS_GUZZLE) {
      // guzzle the gas
      bool usingGas = true;
      while (usingGas) {
        usingGas = true;
      }
    }

    // defaults to EMPTY_REVERT scenario
    revert();
  }
}
