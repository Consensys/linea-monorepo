// SPDX-License-Identifier: MIT-1.0
pragma solidity ^0.8.26;

import { Math } from "@openzeppelin/contracts/utils/math/Math.sol";
import { MultiplierPointMath } from "./MultiplierPointMath.sol";

/**
 * @title StakeMath
 * @author Ricardo Guilherme Schmidt <ricardo3@status.im>
 * @notice Provides mathematical operations and utilities for managing staking operations.
 */
abstract contract StakeMath is MultiplierPointMath {
    error StakeMath__FundsLocked();
    error StakeMath__InvalidLockingPeriod();
    error StakeMath__StakeIsTooLow();
    error StakeMath__InsufficientBalance();
    error StakeMath__AccrueTimeNotReached();
    error StakeMath__AbsoluteMaxMPOverflow();

    event StakeMathTest(uint256 lockTime);
    /// @notice Minimal lockup time

    uint256 public constant MIN_LOCKUP_PERIOD = 90 days;
    /// @notice Maximum lockup period
    uint256 public constant MAX_LOCKUP_PERIOD = MAX_MULTIPLIER * YEAR;

    /**
     * @notice Calculates the bonus multiplier points earned when a balance Δa is increased an optionally locked for a
     * specified duration
     * @param _balance Account current balance
     * @param _currentMaxMP Account current max multiplier points
     * @param _currentLockEndTime Account current lock end timestamp
     * @param _processTime Process current timestamp
     * @param _increasedAmount Increased amount of balance
     * @param _increasedLockSeconds Increased amount of seconds to lock
     * @return _deltaMpTotal Increased amount of total multiplier points
     * @return _deltaMpMax Increased amount of max multiplier points
     * @return _newLockEnd Account new lock end timestamp
     */
    function _calculateStake(
        uint256 _balance,
        uint256 _currentMaxMP,
        uint256 _currentLockEndTime,
        uint256 _processTime,
        uint256 _increasedAmount,
        uint256 _increasedLockSeconds
    )
        internal
        pure
        returns (uint256 _deltaMpTotal, uint256 _deltaMpMax, uint256 _newLockEnd)
    {
        uint256 newBalance = _balance + _increasedAmount;
        _newLockEnd = Math.max(_currentLockEndTime, _processTime) + _increasedLockSeconds;
        // solhint-disable-next-line
        uint256 dtLock = _newLockEnd - _processTime;
        if (dtLock != 0 && (dtLock < MIN_LOCKUP_PERIOD || dtLock > MAX_LOCKUP_PERIOD)) {
            revert StakeMath__InvalidLockingPeriod();
        }

        uint256 deltaMpBonus;
        if (dtLock > 0) {
            deltaMpBonus = _bonusMP(_increasedAmount, dtLock);
        }

        if (_balance > 0 && _increasedLockSeconds > 0) {
            deltaMpBonus += _bonusMP(_balance, _increasedLockSeconds);
        }

        _deltaMpTotal = _initialMP(_increasedAmount) + deltaMpBonus;
        _deltaMpMax = _deltaMpTotal + _accrueMP(_increasedAmount, MAX_MULTIPLIER * YEAR);

        if (_deltaMpMax + _currentMaxMP > MP_MPY_ABSOLUTE * newBalance) {
            revert StakeMath__AbsoluteMaxMPOverflow();
        }
    }

    /**
     * @notice Calculates the bonus multiplier points earned when a balance Δa is locked for a specified duration
     * @param _balance Account current balance
     * @param _currentMaxMP Account current max multiplier points
     * @param _currentLockEndTime Account current lock end timestamp
     * @param _processTime Process current timestamp
     * @param _increasedLockSeconds Increased amount of seconds to lock
     * @return _deltaMp Increased amount of total and max multiplier points
     * @return _newLockEnd Account new lock end timestamp
     */
    function _calculateLock(
        uint256 _balance,
        uint256 _currentMaxMP,
        uint256 _currentLockEndTime,
        uint256 _processTime,
        uint256 _increasedLockSeconds
    )
        internal
        pure
        returns (uint256 _deltaMp, uint256 _newLockEnd)
    {
        if (_balance == 0) {
            revert StakeMath__InsufficientBalance();
        }

        _newLockEnd = Math.max(_currentLockEndTime, _processTime) + _increasedLockSeconds;
        // solhint-disable-next-line
        uint256 dt_lock = _newLockEnd - _processTime;
        if (dt_lock != 0 && (dt_lock < MIN_LOCKUP_PERIOD || dt_lock > MAX_LOCKUP_PERIOD)) {
            revert StakeMath__InvalidLockingPeriod();
        }

        _deltaMp = _bonusMP(_balance, _increasedLockSeconds);

        if (_deltaMp + _currentMaxMP > MP_MPY_ABSOLUTE * _balance) {
            revert StakeMath__AbsoluteMaxMPOverflow();
        }
    }

    /**
     *
     * @param _balance Account current balance
     * @param _currentLockEndTime Account current lock end timestamp
     * @param _processTime Process current timestamp
     * @param _currentTotalMP Account current total multiplier points
     * @param _currentMaxMP Account current max multiplier points
     * @param _reducedAmount Reduced amount of balance
     * @return _deltaMpTotal Increased amount of total multiplier points
     * @return _deltaMpMax Increased amount of max multiplier points
     */
    function _calculateUnstake(
        uint256 _balance,
        uint256 _currentLockEndTime,
        uint256 _processTime,
        uint256 _currentTotalMP,
        uint256 _currentMaxMP,
        uint256 _reducedAmount
    )
        internal
        pure
        returns (uint256 _deltaMpTotal, uint256 _deltaMpMax)
    {
        if (_reducedAmount > _balance) {
            revert StakeMath__InsufficientBalance();
        }
        if (_currentLockEndTime > _processTime) {
            revert StakeMath__FundsLocked();
        }
        _deltaMpTotal = _reduceMP(_balance, _currentTotalMP, _reducedAmount);
        _deltaMpMax = _reduceMP(_balance, _currentMaxMP, _reducedAmount);
    }

    /**
     * @notice Calculates the accrued multiplier points for a given balance and seconds passed since last accrual
     * @param _balance Account current balance
     * @param _currentTotalMP Account current total multiplier points
     * @param _currentMaxMP Account current max multiplier points
     * @param _lastAccrualTime Account current last accrual timestamp
     * @param _processTime Process current timestamp
     * @return _deltaMpTotal Increased amount of total multiplier points
     */
    function _calculateAccrual(
        uint256 _balance,
        uint256 _currentTotalMP,
        uint256 _currentMaxMP,
        uint256 _lastAccrualTime,
        uint256 _processTime
    )
        internal
        pure
        returns (uint256 _deltaMpTotal)
    {
        uint256 dt = _processTime - _lastAccrualTime;
        if (_currentTotalMP < _currentMaxMP) {
            _deltaMpTotal = Math.min(_accrueMP(_balance, dt), _currentMaxMP - _currentTotalMP);
        }
    }

    /**
     * @dev Caution: This value is estimated and can be incorrect due precision loss.
     * @notice Estimates the time an account set as locked time.
     * @param _mpMax Maximum multiplier points calculated from the current balance.
     * @param _balance Current balance used to calculate the maximum multiplier points.
     */
    function _estimateLockTime(uint256 _mpMax, uint256 _balance) internal pure returns (uint256 _lockTime) {
        return Math.mulDiv((_mpMax - _balance) * 100, YEAR, _balance * MP_APY, Math.Rounding.Ceil) - MAX_LOCKUP_PERIOD;
    }
}
