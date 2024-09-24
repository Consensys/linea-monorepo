// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.26;

import { LineaRollup } from "./LineaRollup.sol";

/**
 * @title Contract to reinitialize cross-chain messaging on L1 and rollup proving.
 * @author ConsenSys Software Inc.
 * @dev Init indicates it is an initializer contract
 * @custom:security-contact security-report@linea.build
 */
contract LineaRollupInit is LineaRollup {
  /**
   * @notice Reinitializes LineaRollup and underlying service dependencies.
   * @param _initialStateRootHash The initial hash at migration used for proof verification.
   * @param _initialL2BlockNumber The initial block number at migration.
   */
  function initializeV2(uint256 _initialL2BlockNumber, bytes32 _initialStateRootHash) external reinitializer(3) {
    currentL2BlockNumber = _initialL2BlockNumber;
    stateRootHashes[_initialL2BlockNumber] = _initialStateRootHash;
  }
}
