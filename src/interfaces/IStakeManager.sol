// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { ITrustedCodehashAccess } from "./ITrustedCodehashAccess.sol";
import { IStakeConstants } from "./IStakeConstants.sol";

interface IStakeManager is ITrustedCodehashAccess, IStakeConstants {
    error StakingManager__FundsLocked();
    error StakingManager__InvalidLockTime();
    error StakingManager__InsufficientFunds();
    error StakingManager__StakeIsTooLow();
    error StakingManager__NotAllowedToLeave();

    function registerVault() external;
    function stake(uint256 _amount, uint256 _seconds) external;
    function lock(uint256 _seconds) external;
    function unstake(uint256 _amount) external;
    function leave() external;

    function emergencyModeEnabled() external view returns (bool);
    function totalStaked() external view returns (uint256);
    function totalMPAccrued() external view returns (uint256);
    function totalMaxMP() external view returns (uint256);
    function getStakedBalance(address _vault) external view returns (uint256 _balance);

    function STAKING_TOKEN() external view returns (IERC20);
}
