// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

import { IYieldManager } from "./interfaces/IYieldManager.sol";
import { IYieldProvider } from "./interfaces/IYieldProvider.sol";
import { YieldManagerPauseManager } from "../security/pausing/YieldManagerPauseManager.sol";

/**
 * @title Contract to handle native yield operations.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
contract YieldManager is YieldManagerPauseManager, IYieldManager {
  /// @notice The role required to send ETH to a yield provider.
  bytes32 public constant YIELD_PROVIDER_FUNDER_ROLE = keccak256("YIELD_PROVIDER_FUNDER_ROLE");

  /// @notice The role required to unstake ETH from a yield provider.
  bytes32 public constant YIELD_PROVIDER_UNSTAKER_ROLE = keccak256("YIELD_PROVIDER_UNSTAKER_ROLE");

  /// @notice The role required to request a yield report.
  bytes32 public constant YIELD_REPORTER_ROLE = keccak256("YIELD_REPORTER_ROLE");

  /// @notice The role required to rebalance ETH between the withdrawal reserve and yield provider/s.
  bytes32 public constant RESERVE_OPERATOR_ROLE = keccak256("RESERVE_OPERATOR_ROLE");

  /// @notice The role required to pause beacon chain staking for yield provider/s that support this operation.
  bytes32 public constant STAKING_PAUSER_ROLE = keccak256("STAKING_PAUSER_ROLE");

  /// @notice The role required to unpause beacon chain staking for yield provider/s that support this operation.
  bytes32 public constant UNSTAKING_PAUSER_ROLE = keccak256("UNSTAKING_PAUSER_ROLE");

  /// @notice The role required to execute ossification functions.
  bytes32 public constant OSSIFIER_ROLE = keccak256("OSSIFIER_ROLE");

  /// @custom:storage-location erc7201:linea.storage.YieldManager
  struct YieldManagerStorage {
    address placeholder;
  }

  // keccak256(abi.encode(uint256(keccak256("linea.storage.YieldManagerStorage")) - 1)) & ~bytes32(uint256(0xff))
  bytes32 private constant YieldManagerStorageLocation = 0xdc1272075efdca0b85fb2d76cbb5f26d954dc18e040d6d0b67071bd5cbd04300;

  function _getYieldManagerStorage() private pure returns (YieldManagerStorage storage $) {
      assembly {
          $.slot := YieldManagerStorageLocation
      }
  }

  /**
   * @notice Send ETH to the specified yield strategy.
   * @dev YIELD_PROVIDER_FUNDER_ROLE is required to execute.
   * @dev Reverts if the withdrawal reserve is below the minimum threshold.
   * @dev Will settle any outstanding liabilities to the YieldProvider.
   * @param _amount        The amount of ETH to send.
   * @param _yieldProvider The target yield provider contract.
   */
  function fundYieldProvider(uint256 _amount, address _yieldProvider) external {
    // TODO - Validate withdrawal reserve sufficient
    // TODO - Validate _yieldProvider
    (bool success,) = _yieldProvider.delegatecall(
      abi.encodeCall(IYieldProvider.fundYieldProvider, (_amount)
      ));
     if (!success) {
      revert DelegateCallFailed();
     }
  }

  /**
   * @notice Receive ETH from the withdrawal reserve.
   * @dev Only accepts calls from the withdrawal reserve.
   * @dev Reverts if, after transfer, the withdrawal reserve will be below the minimum threshold.
   */
  function receiveFundsFromReserve() external payable {

  }

  /**
   * @notice Send ETH to the L1MessageService.
   * @dev YIELD_PROVIDER_FUNDER_ROLE or YIELD_MANAGER_UNSTAKER_ROLE is required to execute.
   * @param _amount        The amount of ETH to send.
   */
  function transferFundsToReserve(uint256 _amount) external {

  }

  /**
   * @notice Report newly accrued yield, excluding any portion reserved for system obligations.
   * @dev YIELD_REPORTER_ROLE is required to execute.
   * @dev Since the YieldManager is unaware of donations received via the L1MessageService or L2MessageService,
   *      the `_reserveDonations` parameter is required to ensure accurate yield accounting.
   * @param _totalReserveDonations   Total amount of donations received on the L1MessageService or L2MessageService.
   * @param _yieldProvider      Yield provider address.
   */
  function reportYield(uint256 _totalReserveDonations, address _yieldProvider) external {

  }

  /**
   * @notice Request beacon chain withdrawal from specified yield provider.
   * @dev YIELD_MANAGER_UNSTAKER_ROLE or RESERVE_OPERATOR_ROLE is required to execute.
   * @param _withdrawalParams   Provider-specific withdrawal parameters.
   * @param _yieldProvider      Yield provider address.
   */
  function unstake(bytes memory _withdrawalParams, address _yieldProvider) external {

  }

  /**
   * @notice Permissionlessly request beacon chain withdrawal from a specified yield provider.
   * @dev    Callable only when the withdrawal reserve is in deficit. 
   * @dev    The permissionless unstake amount is capped to the remaining reserve deficit that 
   *         cannot be covered by other liquidity sources:
   *
   *         PERMISSIONLESS_UNSTAKE_AMOUNT ≤
   *           RESERVE_DEFICIT
   *         - YIELD_PROVIDER_BALANCE
   *         - YIELD_MANAGER_BALANCE
   *         - PENDING_PERMISSIONLESS_UNSTAKE
   *
   * @dev Validates (validatorPubkey, validatorBalance, validatorWithdrawalCredential) against EIP-4788 beacon chain root.
   * @param _withdrawalParams       Provider-specific withdrawal parameters.
   * @param _withdrawalParamsProof  Merkle proof of _withdrawalParams to be verified against EIP-4788 beacon chain root.
   * @param _yieldProvider          Yield provider address.
   */
  function unstakePermissionless(
    bytes calldata _withdrawalParams,
    bytes calldata _withdrawalParamsProof,
    address _yieldProvider
  ) external {

  }

  /**
   * @notice Withdraw ETH from a specified yield provider.
   * @dev YIELD_MANAGER_UNSTAKER_ROLE is required to execute.
   * @dev If withdrawal reserve is in deficit, will route funds to the bridge.
   * @dev If fund remaining, will settle any outstanding LST liabilities and protocol obligations.
   * @param _amount                 Amount to withdraw.
   * @param _yieldProvider          Yield provider address.
   */
  function withdrawFromYieldProvider(uint256 _amount, address _yieldProvider) external {

  }

  /**
   * @notice Rebalance ETH from the YieldManager and specified yield provider, sending it to the L1MessageService.
   * @dev RESERVE_OPERATOR_ROLE is required to execute.
   * @dev Settles any outstanding LST liabilities, provided this does not leave the withdrawal reserve in deficit.
   * @param _amount                 Amount to withdraw.
   * @param _yieldProvider          Yield provider address.
   */
  function addToWithdrawalReserve(uint256 _amount, address _yieldProvider) external {

  }

  /**
   * @notice Permissionlessly rebalance ETH from the YieldManager and specified yield provider, sending it to the L1MessageService.
   * @dev Only available when the withdrawal is in deficit.
   * @param _amount                 Amount to withdraw.
   * @param _yieldProvider          Yield provider address.
   */
  function replenishWithdrawalReserve(uint256 _amount, address _yieldProvider) external {

  }

  /**
   * @notice Pauses beacon chain deposits for specified yield provier.
   * @dev STAKING_PAUSER_ROLE is required to execute.
   * @param _yieldProvider          Yield provider address.
   */
  function pauseStaking(address _yieldProvider) external {

  }

  /**
   * @notice Unpauses beacon chain deposits for specified yield provier.
   * @dev STAKING_UNPAUSER_ROLE is required to execute.
   * @dev Will revert if the withdrawal reserve is in deficit, or there is an existing LST liability.
   * @param _yieldProvider          Yield provider address.
   */
  function unpauseStaking(address _yieldProvider) external {

  }
}