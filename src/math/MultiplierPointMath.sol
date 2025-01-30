// SPDX-License-Identifier: MIT-1.0
pragma solidity ^0.8.26;

import { Math } from "@openzeppelin/contracts/utils/math/Math.sol";
import { IStakeConstants } from "../interfaces/IStakeConstants.sol";

/**
 * @title MultiplierPointMath
 * @author Ricardo Guilherme Schmidt <ricardo3@status.im>
 * @notice Provides mathematical operations and utilities for managing multiplier points in the staking system.
 */
abstract contract MultiplierPointMath is IStakeConstants {
    /// @notice One (mean) tropical year, in seconds.
    uint256 public constant YEAR = 365 days;
    /// @notice Accrued multiplier points maximum multiplier.
    uint256 public constant MAX_MULTIPLIER = 4;
    /// @notice Multiplier points annual percentage yield.
    uint256 public constant MP_APY = 100;
    /// @notice Multiplier points accrued maximum percentage yield.
    uint256 public constant MP_MPY = MAX_MULTIPLIER * MP_APY;
    /// @notice Multiplier points absolute maximum percentage yield.
    uint256 public constant MP_MPY_ABSOLUTE = 100 + (2 * (MAX_MULTIPLIER * MP_APY));
    /// @notice The accrue rate period of time over which multiplier points are calculated.
    uint256 public constant ACCRUE_RATE = 1 seconds;
    /// @notice Minimal value to generate 1 multiplier point in the accrue rate period (rounded up).
    uint256 public constant MIN_BALANCE = (((YEAR * 100) - 1) / (MP_APY * ACCRUE_RATE)) + 1;
    /// @notice Maximum value to not overflow unsigned integer of 256 bits.
    uint256 public constant MAX_BALANCE = type(uint256).max / (MP_APY * ACCRUE_RATE);

    /**
     * @notice Calculates the accrued multiplier points (MPs) over a time period Δt, based on the account balance
     * @param _balance Represents the current account balance
     * @param _deltaTime The time difference or the duration over which the multiplier points are accrued, expressed in
     * seconds
     * @return accruedMP points accrued for given `_balance` and  `_seconds`
     */
    function _accrueMP(uint256 _balance, uint256 _deltaTime) internal pure returns (uint256 accruedMP) {
        return Math.mulDiv(_balance, _deltaTime * MP_APY, YEAR * 100);
    }

    /**
     * @notice Calculates the bonus multiplier points (MPs) earned when a balance Δa is locked for a specified duration
     * t_lock.
     * It is equivalent to the accrued multiplier points function but specifically applied in the context of a locked
     * balance.
     * @param _balance quantity of tokens
     * @param _lockedSeconds time in seconds locked
     * @return bonusMP bonus multiplier points for given `_balance` and `_lockedSeconds`
     */
    function _bonusMP(uint256 _balance, uint256 _lockedSeconds) internal pure returns (uint256 bonusMP) {
        return _accrueMP(_balance, _lockedSeconds);
    }

    /**
     * @notice Calculates the initial multiplier points (MPs) based on the balance change Δa. The result is equal to
     * the amount of balance added.
     * @param _balance Represents the change in balance.
     * @return initialMP Initial Multiplier Points
     */
    function _initialMP(uint256 _balance) internal pure returns (uint256 initialMP) {
        return _balance;
    }

    /**
     * @notice Calculates the reduction in multiplier points (MPs) when a portion of the balance Δa `_reducedAmount` is
     * removed from the total balance a_bal `_balance`.
     * The reduction is proportional to the ratio of the removed balance to the total balance, applied to the current
     * multiplier points $mp$.
     * @param _balance The total account balance before the removal of Δa `_reducedBalance`
     * @param _mp Represents the current multiplier points
     * @param _reducedAmount reduced balance
     * @return reducedMP Multiplier points to reduce from `_mp`
     */
    function _reduceMP(
        uint256 _balance,
        uint256 _mp,
        uint256 _reducedAmount
    )
        internal
        pure
        returns (uint256 reducedMP)
    {
        return Math.mulDiv(_mp, _reducedAmount, _balance);
    }

    /**
     * @notice Calculates maximum stake a given `_balance` can be generated with `MAX_MULTIPLIER`
     * @param _balance quantity of tokens
     * @return maxMPAccrued maximum quantity of muliplier points that can be generated for given `_balance`
     */
    function _maxAccrueMP(uint256 _balance) internal pure returns (uint256 maxMPAccrued) {
        return Math.mulDiv(_balance, MP_MPY, 100);
    }

    /**
     * @notice The maximum total multiplier points that can be generated for a determined amount of balance and lock
     * duration.
     * @param _balance Represents the current account balance
     * @param _lockTime The time duration for which the balance is locked
     * @return maxMP Maximum Multiplier Points that can be generated for given `_balance` and `_lockTime`
     */
    function _maxTotalMP(uint256 _balance, uint256 _lockTime) internal pure returns (uint256 maxMP) {
        return _balance + Math.mulDiv(_balance * MP_APY, (MAX_MULTIPLIER * YEAR) + _lockTime, YEAR * 100);
    }

    /**
     * @notice The absolute maximum total multiplier points that some balance could have, which is the sum of the
     * maximum
     * lockup time bonus possible and the maximum accrued multiplier points.
     * @param _balance quantity of tokens
     * @return maxMPAbsolute Absolute Maximum Multiplier Points
     */
    function _maxAbsoluteTotalMP(uint256 _balance) internal pure returns (uint256 maxMPAbsolute) {
        return Math.mulDiv(_balance, MP_MPY_ABSOLUTE, 100);
    }

    /**
     * @dev Caution: This value is estimated and can be incorrect due precision loss.
     * @notice Calculates the remaining lock time available for a given `_mpMax` and `_balance`
     * @param _balance Current balance used to calculate the maximum multiplier points.
     * @param _mpMax Maximum multiplier points calculated from the current balance.
     * @return lockTime Amount of lock time allowed to be increased
     */
    function _lockTimeAvailable(uint256 _balance, uint256 _mpMax) internal pure returns (uint256 lockTime) {
        return Math.mulDiv((_balance * MP_MPY_ABSOLUTE) - _mpMax, YEAR, _balance * 100);
    }

    /**
     * @notice Calculates the time required to accrue a specific multiplier point value.
     * @param _balance The current balance.
     * @param _mp The target multiplier points to accrue.
     * @return timeToReachMaxMP The time required to reach the specified multiplier points, in seconds.
     */
    function _timeToAccrueMP(uint256 _balance, uint256 _mp) internal pure returns (uint256 timeToReachMaxMP) {
        return Math.mulDiv(_mp * 100, YEAR, _balance * MP_APY);
    }

    /**
     * @notice Calculates the bonus multiplier points based on the balance and maximum multiplier points.
     * @param _balance The current balance.
     * @param _maxMP The maximum multiplier points.
     * @return bonusMP The calculated bonus multiplier points.
     */
    function _retrieveBonusMP(uint256 _balance, uint256 _maxMP) internal pure returns (uint256 bonusMP) {
        return _maxMP - (_balance + _maxAccrueMP(_balance));
    }

    /**
     * @notice Retrieves the accrued multiplier points based on the total and maximum multiplier points.
     * @param _balance The current balance.
     * @param _totalMP The total multiplier points.
     * @param _maxMP The maximum multiplier points.
     * @return accruedMP The calculated accrued multiplier points.
     */
    function _retrieveAccruedMP(
        uint256 _balance,
        uint256 _totalMP,
        uint256 _maxMP
    )
        internal
        pure
        returns (uint256 accruedMP)
    {
        return _totalMP + _maxAccrueMP(_balance) - _maxMP;
    }
}
