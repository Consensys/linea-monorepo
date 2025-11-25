// SPDX-License-Identifier: AGPL-3.0
pragma solidity ^0.8.30;

import { LineaRollupBase } from "./LineaRollupBase.sol";
import { IFallbackOperator } from "./interfaces/IFallbackOperator.sol";
/**
 * @title Contract to manage EIP-4844 blob submission.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
abstract contract FallbackOperator is LineaRollupBase, IFallbackOperator {
  /// @dev In practice, when used, this is expected to be a close approximation to 6 months, and is intentional.
  uint256 internal constant SIX_MONTHS_IN_SECONDS = (365 / 2) * 24 * 60 * 60;

  /// @dev Keep 50 free storage slots for inheriting contracts.
  uint256[50] private __gap_FallbackOperator;

  /**
   * @notice Initializes _FallbackOperator - used for new networks only.
   * @param _fallbackOperator The initial computed genesis shnarf.
   */
  function __FallbackOperator_init(address _fallbackOperator) internal virtual {
    require(_fallbackOperator != address(0), ZeroAddressNotAllowed());
    fallbackOperator = _fallbackOperator;
    emit FallbackOperatorAddressSet(msg.sender, _fallbackOperator);
  }

  /**
   * @notice Sets the fallback operator role to the specified address if six months have passed since the last finalization.
   * @dev Reverts if six months have not passed since the last finalization.
   * @param _messageNumber Last finalized L1 message number as part of the feedback loop.
   * @param _rollingHash Last finalized L1 rolling hash as part of the feedback loop.
   * @param _lastFinalizedTimestamp Last finalized L2 block timestamp.
   */
  function setFallbackOperator(uint256 _messageNumber, bytes32 _rollingHash, uint256 _lastFinalizedTimestamp) external {
    if (block.timestamp < _lastFinalizedTimestamp + SIX_MONTHS_IN_SECONDS) {
      revert LastFinalizationTimeNotLapsed();
    }
    if (currentFinalizedState != _computeLastFinalizedState(_messageNumber, _rollingHash, _lastFinalizedTimestamp)) {
      revert FinalizationStateIncorrect(
        currentFinalizedState,
        _computeLastFinalizedState(_messageNumber, _rollingHash, _lastFinalizedTimestamp)
      );
    }

    address fallbackOperatorAddress = fallbackOperator;

    _grantRole(OPERATOR_ROLE, fallbackOperatorAddress);
    emit FallbackOperatorRoleGranted(msg.sender, fallbackOperatorAddress);
  }

  /**
   * @notice Revokes `role` from the calling account.
   * @dev Fallback operator cannot renounce role. Reverts with OnlyNonFallbackOperator.
   * @param _role The role to renounce.
   * @param _account The account to renounce - can only be the _msgSender().
   */
  function renounceRole(bytes32 _role, address _account) public virtual override {
    if (_account == fallbackOperator) {
      revert OnlyNonFallbackOperator();
    }

    super.renounceRole(_role, _account);
  }
}
