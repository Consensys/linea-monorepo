// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { IStakeManager } from "./../../src/interfaces/IStakeManager.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { TrustedCodehashAccess } from "./../../src/TrustedCodehashAccess.sol";
import { UUPSUpgradeable } from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import { ReentrancyGuardUpgradeable } from "@openzeppelin/contracts-upgradeable/utils/ReentrancyGuardUpgradeable.sol";

contract StackOverflowStakeManager is
    UUPSUpgradeable,
    IStakeManager,
    TrustedCodehashAccess,
    ReentrancyGuardUpgradeable
{
    IERC20 public STAKING_TOKEN;

    uint256 public constant SCALE_FACTOR = 1e18;
    uint256 public constant MP_RATE_PER_YEAR = 1e18;

    uint256 public constant MIN_LOCKUP_PERIOD = 90 days;
    uint256 public constant MAX_LOCKUP_PERIOD = 4 * 365 days;
    uint256 public constant MAX_MULTIPLIER = 4;

    uint256 public totalStaked;
    uint256 public totalMPAccrued;
    uint256 public totalMaxMP;
    uint256 public rewardIndex;
    uint256 public accountedRewards;
    uint256 public lastMPUpdatedTime;
    bool public emergencyModeEnabled;

    struct Account {
        uint256 stakedBalance;
        uint256 accountRewardIndex;
        uint256 accountMP;
        uint256 maxMP;
        uint256 lastMPUpdateTime;
        uint256 lockUntil;
    }

    mapping(address account => Account data) public accounts;

    function getStakedBalance(address _vault) external view override returns (uint256 _balance) {
        // implementation
    }
    function lock(uint256 _seconds) external override {
        // implementation
    }
    function stake(uint256 _amount, uint256 _seconds) external override {
        // implementation
    }
    function unstake(uint256 _amount) external override {
        // implementation
    }

    function leave() external override {
        this.leave();
    }

    function _authorizeUpgrade(address) internal view override {
        _checkOwner();
    }

    function getAccount(address _account) external view returns (Account memory) {
        return accounts[_account];
    }

    function registerVault() external override {
        // implementation
    }
}
