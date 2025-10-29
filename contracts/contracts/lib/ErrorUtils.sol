// SPDX-License-Identifier: AGPL-3.0
pragma solidity ^0.8.30;

import { IGenericErrors } from "../interfaces/IGenericErrors.sol";

/**
 * @title Library to perform SparseMerkleProof actions using the MiMC hashing algorithm
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
library ErrorUtils {
  function revertIfZeroAddress(address _addr) internal pure {
    if (_addr == address(0)) revert IGenericErrors.ZeroAddressNotAllowed();
  }
}
