// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

/**
 * @title IStakeConstants
 * @author Ricardo Guilherme Schmidt <ricardo3@status.im>
 * @notice Interface for Stake Constants
 * @dev This interface is necessary to linearize the inheritance of StakeMath and MultiplierPointMath
 */
interface IStakeConstants {
    function MIN_LOCKUP_PERIOD() external view returns (uint256);
    function MAX_LOCKUP_PERIOD() external view returns (uint256);
    function MP_APY() external view returns (uint256);
    function MAX_MULTIPLIER() external view returns (uint256);
}
