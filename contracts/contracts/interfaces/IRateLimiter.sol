// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.19;

interface IRateLimiter {
  /**
   * @dev Thrown when an amount breaches the limit in the period.
   */
  error RateLimitExceeded();

  /**
   * @dev Thrown when the period is initialised to zero.
   */
  error PeriodIsZero();

  /**
   * @dev Thrown when the limit is initialised to zero.
   */
  error LimitIsZero();

  /**
   * @dev Emitted when the amount in the period is reset to zero.
   */
  event AmountUsedInPeriodReset(address indexed resettingAddress);

  /**
   * @dev Emitted when the limit is changed.
   * @dev If the current used amount is higher than the new limit, the used amount is lowered to the limit.
   */
  event LimitAmountChanged(
    address indexed amountChangeBy,
    uint256 amount,
    bool amountUsedLoweredToLimit,
    bool usedAmountResetToZero
  );

  /**
   * @notice Resets the rate limit amount to the amount specified.
   * @param _amount New message hashes.
   */
  function resetRateLimitAmount(uint256 _amount) external;

  /**
   * @notice Resets the amount used in the period to zero.
   */
  function resetAmountUsedInPeriod() external;
}
