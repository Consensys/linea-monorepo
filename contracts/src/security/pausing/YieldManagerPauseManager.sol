// SPDX-License-Identifier: AGPL-3.0
pragma solidity ^0.8.30;

import { PauseManager } from "./PauseManager.sol";

/**
 * @title Contract to manage pausing roles for the YieldManager.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
abstract contract YieldManagerPauseManager is PauseManager {
      /// @notice This is used to pause token bridging initiation.
  bytes32 public constant PAUSE_INITIATE_TOKEN_BRIDGING_ROLE = keccak256("PAUSE_INITIATE_TOKEN_BRIDGING_ROLE");

  /// @notice This is used to unpause token bridging initiation.
  bytes32 public constant UNPAUSE_INITIATE_TOKEN_BRIDGING_ROLE = keccak256("UNPAUSE_INITIATE_TOKEN_BRIDGING_ROLE");

    // NATIVE_YIELD_STAKING,
    // NATIVE_YIELD_UNSTAKING,
    // NATIVE_YIELD_PERMISSIONLESS_UNSTAKING,
    // NATIVE_YIELD_PERMISSIONLESS_REBALANCE,
    // NATIVE_YIELD_RESERVE_FUNDING,
    // NATIVE_YIELD_REPORTING,
    // LST_WITHDRAWAL
}
