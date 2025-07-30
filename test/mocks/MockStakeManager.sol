pragma solidity 0.8.26;

import { IStakeManager } from "../../src/interfaces/IStakeManager.sol";
import { ITrustedCodehashAccess } from "../../src/interfaces/ITrustedCodehashAccess.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract MockStakeManager is ITrustedCodehashAccess, IStakeManager {
    function implementation() external view returns (address) {
        return address(this);
    }

    function setTrustedCodehash(bytes32, bool) external {
        return;
    }

    function isTrustedCodehash(bytes32) external view returns (bool) {
        return true;
    }

    function registerVault() external {
        return;
    }

    function stake(uint256, uint256) external {
        return;
    }

    function lock(uint256) external {
        return;
    }

    function unstake(uint256) external {
        return;
    }

    function leave() external {
        return;
    }

    function migrateToVault(address) external {
        return;
    }

    function updateVault(address) external {
        return;
    }

    function emergencyModeEnabled() external view returns (bool) {
        return false;
    }

    function totalStaked() external view returns (uint256) {
        return 0;
    }

    function totalMPAccrued() external view returns (uint256) {
        return 0;
    }

    function totalMaxMP() external view returns (uint256) {
        return 0;
    }

    function stakedBalanceOf(address) external view returns (uint256) {
        return 0;
    }

    function STAKING_TOKEN() external view returns (IERC20) {
        return IERC20(address(0));
    }

    function MIN_LOCKUP_PERIOD() external view returns (uint256) {
        return 0;
    }

    function MAX_LOCKUP_PERIOD() external view returns (uint256) {
        return 0;
    }

    function MP_APY() external view returns (uint256) {
        return 0;
    }

    function MAX_MULTIPLIER() external view returns (uint256) {
        return 0;
    }
}
