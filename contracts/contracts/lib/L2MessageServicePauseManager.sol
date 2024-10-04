// SPDX-License-Identifier: AGPL-3.0
pragma solidity >=0.8.19 <=0.8.26;

import { PauseManager } from "./PauseManager.sol";

/**
 * @title Contract to manage cross-chain function pausing for LineaRollup.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
abstract contract L2MessageServicePauseManager is PauseManager {
  bytes32 public constant PAUSE_L1_L2_ROLE = keccak256("PAUSE_L1_L2_ROLE");
  bytes32 public constant UNPAUSE_L1_L2_ROLE = keccak256("UNPAUSE_L1_L2_ROLE");
  bytes32 public constant PAUSE_L2_L1_ROLE = keccak256("PAUSE_L2_L1_ROLE");
  bytes32 public constant UNPAUSE_L2_L1_ROLE = keccak256("UNPAUSE_L2_L1_ROLE");
}
