// SPDX-License-Identifier: AGPL-3.0
pragma solidity ^0.8.30;

import { PauseManager } from "./PauseManager.sol";

/**
 * @title Contract to manage pausing roles for the YieldManager.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
abstract contract YieldManagerPauseManager is PauseManager {
  /// @notice This is used to pause native-yield driven funding of external strategies.
  bytes32 public constant PAUSE_NATIVE_YIELD_STAKING_ROLE = keccak256("PAUSE_NATIVE_YIELD_STAKING_ROLE");

  /// @notice This is used to unpause native-yield driven funding of external strategies.
  bytes32 public constant UNPAUSE_NATIVE_YIELD_STAKING_ROLE = keccak256("UNPAUSE_NATIVE_YIELD_STAKING_ROLE");

  /// @notice This is used to pause operator-led unstaking flows.
  bytes32 public constant PAUSE_NATIVE_YIELD_UNSTAKING_ROLE = keccak256("PAUSE_NATIVE_YIELD_UNSTAKING_ROLE");

  /// @notice This is used to unpause operator-led unstaking flows.
  bytes32 public constant UNPAUSE_NATIVE_YIELD_UNSTAKING_ROLE = keccak256("UNPAUSE_NATIVE_YIELD_UNSTAKING_ROLE");

  /// @notice This is used to pause permissionless unstaking requests.
  bytes32 public constant PAUSE_NATIVE_YIELD_PERMISSIONLESS_UNSTAKING_ROLE =
    keccak256("PAUSE_NATIVE_YIELD_PERMISSIONLESS_UNSTAKING_ROLE");

  /// @notice This is used to unpause permissionless unstaking requests.
  bytes32 public constant UNPAUSE_NATIVE_YIELD_PERMISSIONLESS_UNSTAKING_ROLE =
    keccak256("UNPAUSE_NATIVE_YIELD_PERMISSIONLESS_UNSTAKING_ROLE");

  /// @notice This is used to pause permissionless reserve replenishment flows.
  bytes32 public constant PAUSE_NATIVE_YIELD_PERMISSIONLESS_REBALANCE_ROLE =
    keccak256("PAUSE_NATIVE_YIELD_PERMISSIONLESS_REBALANCE_ROLE");

  /// @notice This is used to unpause permissionless reserve replenishment flows.
  bytes32 public constant UNPAUSE_NATIVE_YIELD_PERMISSIONLESS_REBALANCE_ROLE =
    keccak256("UNPAUSE_NATIVE_YIELD_PERMISSIONLESS_REBALANCE_ROLE");

  /// @notice This is used to pause transfers from the YieldManager to the withdrawal reserve.
  bytes32 public constant PAUSE_NATIVE_YIELD_RESERVE_FUNDING_ROLE =
    keccak256("PAUSE_NATIVE_YIELD_RESERVE_FUNDING_ROLE");

  /// @notice This is used to unpause transfers from the YieldManager to the withdrawal reserve.
  bytes32 public constant UNPAUSE_NATIVE_YIELD_RESERVE_FUNDING_ROLE =
    keccak256("UNPAUSE_NATIVE_YIELD_RESERVE_FUNDING_ROLE");

  /// @notice This is used to pause yield-reporting operations.
  bytes32 public constant PAUSE_NATIVE_YIELD_REPORTING_ROLE = keccak256("PAUSE_NATIVE_YIELD_REPORTING_ROLE");

  /// @notice This is used to unpause yield-reporting operations.
  bytes32 public constant UNPAUSE_NATIVE_YIELD_REPORTING_ROLE = keccak256("UNPAUSE_NATIVE_YIELD_REPORTING_ROLE");

  /// @notice This is used to pause LST withdrawals routed through the YieldManager.
  bytes32 public constant PAUSE_LST_WITHDRAWAL_ROLE = keccak256("PAUSE_LST_WITHDRAWAL_ROLE");

  /// @notice This is used to unpause LST withdrawals routed through the YieldManager.
  bytes32 public constant UNPAUSE_LST_WITHDRAWAL_ROLE = keccak256("UNPAUSE_LST_WITHDRAWAL_ROLE");
}
